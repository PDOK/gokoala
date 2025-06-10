package features

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/PDOK/gokoala/internal/ogc/features/domain"
)

const projInfoTool = "projinfo"

var (
	execCommand  = exec.Command  // Allow mocking
	execLookPath = exec.LookPath // Allow mocking
)

// ProjInfo output in PROJJSON format. Note: only relevant fields are mapped in this struct.
type ProjInfo struct {
	CoordinateSystem CoordinateSystem `json:"coordinate_system"` //nolint:tagliatelle
}

// CoordinateSystem represents the CRS definition
type CoordinateSystem struct {
	Axis []Axis `json:"axis"`
}

// Axis represents a CRS axis
type Axis struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	Direction    string `json:"direction"`
	Unit         string `json:"unit"`
}

// ShouldSwapXY true when given SRID should be YX, false when SRID should be XY.
func ShouldSwapXY(srid domain.SRID) (bool, error) {
	epsgCode := fmt.Sprintf("EPSG:%d", srid)
	info, err := execProjInfo(epsgCode)
	if err != nil {
		return false, err
	}
	// east/north == XY, north/east == YX.
	return info.CoordinateSystem.Axis[0].Direction == "north", nil
}

func execProjInfo(epsgCode string) (*ProjInfo, error) {
	_, err := execLookPath(projInfoTool)
	if err != nil {
		return nil, fmt.Errorf("%s command not found in PATH: %w", projInfoTool, err)
	}

	// Run 'projinfo' and return output in PROJJSON format (https://proj.org/en/stable/specifications/projjson.html)
	cmd := execCommand(projInfoTool, epsgCode, "-o", "projjson", "--single-line", "-q")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute %s command: %w", projInfoTool, err)
	}

	var projInfo ProjInfo
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(output))), &projInfo); err != nil {
		return nil, fmt.Errorf("failed to parse %s output: %w", projInfoTool, err)
	}
	if len(projInfo.CoordinateSystem.Axis) < 1 {
		return nil, fmt.Errorf("invalid %s output: axis not found", projInfoTool)
	}
	return &projInfo, nil
}
