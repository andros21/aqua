package cp

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (c *Controller) install(ctx context.Context, logE *logrus.Entry, findResult *which.FindResult, policyConfigs []*policy.Config) error {
	var checksums *checksum.Checksums
	if findResult.Config.ChecksumEnabled() {
		checksums = checksum.New()
		checksumFilePath, err := checksum.GetChecksumFilePathFromConfigFilePath(c.fs, findResult.ConfigFilePath)
		if err != nil {
			return err //nolint:wrapcheck
		}
		if err := checksums.ReadFile(c.fs, checksumFilePath); err != nil {
			return fmt.Errorf("read a checksum JSON: %w", err)
		}
		defer func() {
			if err := checksums.UpdateFile(c.fs, checksumFilePath); err != nil {
				logE.WithError(err).Error("update a checksum file")
			}
		}()
	}

	if err := c.packageInstaller.InstallPackage(ctx, logE, &installpackage.ParamInstallPackage{
		Pkg:             findResult.Package,
		Checksums:       checksums,
		RequireChecksum: findResult.Config.RequireChecksum(c.requireChecksum),
		ConfigFileDir:   filepath.Dir(findResult.ConfigFilePath),
		PolicyConfigs:   policyConfigs,
	}); err != nil {
		return fmt.Errorf("install a package: %w", logerr.WithFields(err, logE.Data))
	}
	return c.packageInstaller.WaitExe(ctx, logE, findResult.ExePath) //nolint:wrapcheck
}
