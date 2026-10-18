//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/tc-hib/winres"
	"github.com/tc-hib/winres/version"
)

func main() {
	rs := &winres.ResourceSet{}

	manifest := winres.AppManifest{
		Identity: winres.AssemblyIdentity{
			Name:    "com.wails.lanchat",
			Version: [4]uint16{1, 0, 0, 0},
		},
		Description:         "LAN Chat and File Transfer Tool",
		Compatibility:       winres.Win7AndAbove,
		ExecutionLevel:      winres.AsInvoker,
		DPIAwareness:        winres.DPIPerMonitorV2,
		UseCommonControlsV6: true,
	}
	rs.SetManifest(manifest)

	iconData, err := os.ReadFile("build/windows/icon.ico")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read icon: %v\n", err)
		os.Exit(1)
	}

	icon, err := winres.LoadICO(bytes.NewReader(iconData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "load icon: %v\n", err)
		os.Exit(1)
	}

	if err := rs.SetIcon(winres.ID(3), icon); err != nil {
		fmt.Fprintf(os.Stderr, "set icon: %v\n", err)
		os.Exit(1)
	}

	vi := version.Info{}
	vi.FileVersion = [4]uint16{1, 0, 0, 0}
	vi.ProductVersion = [4]uint16{1, 0, 0, 0}
	vi.Set(0x0409, version.ProductName, "LAN Chat")
	vi.Set(0x0409, version.ProductVersion, "1.0.0")
	vi.Set(0x0409, version.FileVersion, "1.0.0")
	vi.Set(0x0409, version.FileDescription, "LAN Chat and File Transfer Tool")
	vi.Set(0x0409, version.LegalCopyright, "Copyright 2026")
	vi.Set(0x0409, version.CompanyName, "LAN Chat")
	rs.SetVersionInfo(vi)

	outPath := "rsrc_windows_amd64.syso"
	f, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create syso: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := rs.WriteObject(f, winres.ArchAMD64); err != nil {
		fmt.Fprintf(os.Stderr, "write syso: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated %s\n", outPath)
}
