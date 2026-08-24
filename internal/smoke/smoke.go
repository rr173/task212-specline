// Package smoke 实现 --smoke-test 端到端自检：真实导入光谱峰、估计波长漂移、
// 生成互斥归属候选、标记宇宙线伪线、登记反证、确认归属并冻结版本，
// 关闭并重新打开数据库验证持久化与重启恢复，最后以 0 退出码结束。
package smoke

import (
	"fmt"
	"os"

	"task212-specline/internal/model"
	"task212-specline/internal/observation"
	"task212-specline/internal/service"
	"task212-specline/internal/store"
)

// Main 自检入口：args[0] 为数据库路径。
func Main(args []string) {
	dbPath := "specline.db"
	if len(args) > 0 && args[0] != "" {
		dbPath = args[0]
	}
	if err := Run(dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "smoke test FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("smoke test PASSED")
}

// Run 执行端到端自检。返回 nil 表示全部通过。
func Run(dbPath string) error {
	var obsID, frozenVersionID int64

	step := func(n int, name string, fn func() error) error {
		fmt.Printf("[smoke %d/8] %s ...\n", n, name)
		if err := fn(); err != nil {
			return fmt.Errorf("smoke step %d (%s): %w", n, name, err)
		}
		return nil
	}

	// 第一步：导入观测与光谱峰（含一条宇宙线伪峰、一条互斥双线区峰）。
	if err := step(1, "导入观测与光谱峰", func() error {
		db, err := store.Open(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()
		app := service.New(db)

		o, err := app.Observation.Create("smoke-spectrum", "HD-123456", model.UnitAngstrom)
		if err != nil {
			return err
		}
		obsID = o.ID
		_, err = app.Observation.AddPeaks(obsID, []observation.PeakInput{
			{Index: 0, Wavelength: 3727.70, Unit: model.UnitAngstrom, Flux: 0.8, Region: "blue"},
			{Index: 1, Wavelength: 4700.30, Unit: model.UnitAngstrom, Flux: 9.9, Region: "mid"}, // 宇宙线伪峰
			{Index: 2, Wavelength: 4861.63, Unit: model.UnitAngstrom, Flux: 3.1, Region: "mid"},
			{Index: 3, Wavelength: 6563.09, Unit: model.UnitAngstrom, Flux: 6.4, Region: "red"},
		})
		return err
	}); err != nil {
		return err
	}

	// 第二步：估计波长漂移并校准。
	if err := step(2, "波长漂移校准", func() error {
		db, err := store.Open(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()
		app := service.New(db)
		c, err := app.Calibration.Calibrate(obsID, model.CalibModelOffset)
		if err != nil {
			return err
		}
		// 漂移应接近 +0.30 Å（H-alpha/H-beta 两条参考线）。
		if c.ReferenceCount < 2 {
			return fmt.Errorf("expected >=2 reference lines, got %d", c.ReferenceCount)
		}
		if c.Offset < 0.25 || c.Offset > 0.35 {
			return fmt.Errorf("unexpected offset %.4f", c.Offset)
		}
		return nil
	}); err != nil {
		return err
	}

	// 第三步：生成归属候选并识别互斥。
	var exclusiveCandIDs []int64
	if err := step(3, "生成归属候选并识别互斥", func() error {
		db, err := store.Open(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()
		app := service.New(db)
		cands, err := app.Matching.Match(obsID, 2.0)
		if err != nil {
			return err
		}
		for _, c := range cands {
			if c.Status == model.CandidateMutuallyExclusive {
				exclusiveCandIDs = append(exclusiveCandIDs, c.ID)
			}
		}
		if len(exclusiveCandIDs) != 2 {
			return fmt.Errorf("expected 2 mutually-exclusive candidates, got %d", len(exclusiveCandIDs))
		}
		return nil
	}); err != nil {
		return err
	}

	// 第四步：标记并排除宇宙线伪峰。
	if err := step(4, "标记并排除宇宙线伪峰", func() error {
		db, err := store.Open(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()
		app := service.New(db)
		peaks, err := app.Observation.GetPeaks(obsID)
		if err != nil {
			return err
		}
		// 找到 4700 Å 附近的伪峰（index 1）。
		var artifactPeakID int64
		for _, p := range peaks {
			if p.Index == 1 {
				artifactPeakID = p.ID
			}
		}
		if _, err := app.Review.MarkCosmicRay(artifactPeakID); err != nil {
			return err
		}
		excluded, err := app.Review.ExcludePeak(artifactPeakID)
		if err != nil {
			return err
		}
		if excluded.Status != model.PeakExcluded {
			return fmt.Errorf("artifact peak not excluded")
		}
		return nil
	}); err != nil {
		return err
	}

	// 第五步：登记反证并确认一条归属。
	if err := step(5, "登记反证并确认归属", func() error {
		db, err := store.Open(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()
		app := service.New(db)
		// 互斥候选按 transition_key 区分：否决 OII-3726，确认 OII-3729。
		var refuteID, confirmID int64
		for _, cid := range exclusiveCandIDs {
			c, err := app.Matching.Candidate(cid)
			if err != nil {
				return err
			}
			if c.TransitionKey == "OII-3726" {
				refuteID = cid
			} else if c.TransitionKey == "OII-3729" {
				confirmID = cid
			}
		}
		if refuteID == 0 || confirmID == 0 {
			return fmt.Errorf("failed to locate OII doublet candidates")
		}
		if _, err := app.Review.AddRefutation(refuteID, "velocity", "该峰红移速度与 OII-3726 不符"); err != nil {
			return err
		}
		confirmed, err := app.Matching.Confirm(confirmID)
		if err != nil {
			return err
		}
		if confirmed.Status != model.CandidateSufficient {
			return fmt.Errorf("candidate not sufficient")
		}
		return nil
	}); err != nil {
		return err
	}

	// 第六步：冻结归属版本。
	if err := step(6, "冻结归属版本", func() error {
		db, err := store.Open(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()
		app := service.New(db)
		v, err := app.Versioning.Create(obsID, "smoke-v1")
		if err != nil {
			return err
		}
		frozen, err := app.Versioning.Freeze(v.ID)
		if err != nil {
			return err
		}
		if frozen.Status != model.VersionFrozen {
			return fmt.Errorf("version not frozen")
		}
		frozenVersionID = frozen.ID
		return nil
	}); err != nil {
		return err
	}

	// 第七步：关闭并重开数据库验证持久化与重启恢复。
	if err := step(7, "关闭并重开数据库验证恢复", func() error {
		db, err := store.Open(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()
		app := service.New(db)

		o, err := app.DB.GetObservation(obsID)
		if err != nil {
			return err
		}
		peaks, err := app.Observation.GetPeaks(obsID)
		if err != nil {
			return err
		}
		if len(peaks) != 4 {
			return fmt.Errorf("expected 4 peaks after reopen, got %d", len(peaks))
		}
		v, err := app.Versioning.Get(frozenVersionID)
		if err != nil {
			return err
		}
		if v.Status != model.VersionFrozen {
			return fmt.Errorf("version lost frozen status after reopen")
		}
		if o.ContentHash == "" {
			return fmt.Errorf("content hash missing after reopen")
		}
		return nil
	}); err != nil {
		return err
	}

	// 第八步：统计校验。
	if err := step(8, "统计校验", func() error {
		db, err := store.Open(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()
		app := service.New(db)
		st, err := app.Stats()
		if err != nil {
			return err
		}
		if st.Observations < 1 || st.Peaks < 4 || st.Candidates < 4 || st.Versions < 1 {
			return fmt.Errorf("unexpected stats: %+v", st)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
