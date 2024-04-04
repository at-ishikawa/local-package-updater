package packagemanager

import (
	"fmt"
	"log/slog"
	"strings"
)

type Plugin interface {
	IsSudoRequired() bool
	IsCommandInstalled() bool
	Update() error
}

type GeneralManager struct {
	cliArgs        CLIArgs
	isSudoRequired bool
}

var (
	_ Plugin = (*GeneralManager)(nil)
)

func NewGeneralManager(args CLIArgs, isSudoRequired bool) GeneralManager {
	return GeneralManager{
		cliArgs:        args,
		isSudoRequired: isSudoRequired,
	}
}

func (gpm GeneralManager) IsSudoRequired() bool {
	return gpm.isSudoRequired
}

func (gpm GeneralManager) IsCommandInstalled() bool {
	return checkCommandExists(gpm.cliArgs[0])
}

func (gpm GeneralManager) Update() error {
	_, _, err := runCommand(gpm.cliArgs)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}
	return nil
}

type AptManager struct {
	isSudoRequired bool
}

var (
	_ Plugin = (*AptManager)(nil)
)

func NewAptManager(isSudoRequired bool) *AptManager {
	return &AptManager{
		isSudoRequired: isSudoRequired,
	}
}

func (aptManager AptManager) IsSudoRequired() bool {
	return aptManager.isSudoRequired
}

func (aptManager AptManager) IsCommandInstalled() bool {
	return checkCommandExists("apt")
}

func (aptManager AptManager) ListUpgradablePackages() ([]string, error) {
	args := CLIArgs{"apt", "list", "--upgradable", "--manual-installed"}
	stdout, _, err := runCommand(args)
	if err != nil {
		return nil, err
	}

	// for some reasons, if there is no element, it'll be [\n]
	stdout = strings.TrimSpace(stdout)
	if len(stdout) == 0 {
		return nil, nil
	}

	var result []string
	// The first line is always be Listing...
	for _, line := range strings.Split(stdout, "\n")[1:] {
		result = append(result, strings.Split(line, "/")[0])
	}
	return result, nil
}

func (aptManager AptManager) UpgradeAllPackages(packages []string) (string, error) {
	args := append(CLIArgs{"sudo", "apt", "upgrade", "--only-upgrade", "--yes"}, packages...)
	stdout, stderr, err := runCommand(args)
	if err != nil {
		return stderr, err
	}
	return stdout, nil
}

func (aptManager AptManager) updatePackageList() (string, error) {
	args := CLIArgs{"sudo", "apt", "update"}
	stdout, stderr, err := runCommand(args)
	if err != nil {
		return stderr, err
	}
	return stdout, nil
}

func (aptManager AptManager) removeUnusedPackages() (string, error) {
	stdout, stderr, err := runCommand(CLIArgs{"sudo", "apt", "autoremove"})
	if err != nil {
		return stderr, err
	}
	return stdout, nil
}

func (aptManager AptManager) Update() error {
	_, err := aptManager.removeUnusedPackages()
	if err != nil {
		return err
	}

	_, err = aptManager.updatePackageList()
	if err != nil {
		return err
	}

	packages, err := aptManager.ListUpgradablePackages()
	if err != nil {
		return err
	}
	slog.Debug("apt upgradable packages",
		slog.Any("packages", packages))
	if len(packages) == 0 {
		return nil
	}
	_, err = aptManager.UpgradeAllPackages(packages)
	if err != nil {
		return err
	}

	return nil
}
