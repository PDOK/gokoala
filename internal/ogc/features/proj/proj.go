package proj

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

// Info output in PROJJSON format. Note: only relevant fields are mapped in this struct.
type Info struct {
	CoordinateSystem struct {
		Axis []struct {
			Name         string `json:"name"`
			Abbreviation string `json:"abbreviation"`
			Direction    string `json:"direction"`
			Unit         string `json:"unit"`
		} `json:"axis"`
	} `json:"coordinate_system"` //nolint:tagliatelle
}

// GetAxisOrderForCRS returns XY or YX axis order for the given SRID.
// epsg must be a CRS identifier such as "EPSG:28992" or "OGC:CRS84"
func GetAxisOrderForCRS(epsg string) (domain.AxisOrder, error) {
	projJSON, err := ExecProjInfo(epsg)
	if err != nil {
		return domain.AxisOrderXY, fmt.Errorf("get CRS information for %q: %w", epsg, err)
	}
	axisOrder, err := GetAxisOrder(projJSON)
	if err != nil {
		return domain.AxisOrderXY, fmt.Errorf("get axis order for %q: %w", epsg, err)
	}
	return axisOrder, nil
}

// GetAxisOrder returns XY or YX axis order for the given SRID in PROJJSON format.
func GetAxisOrder(projJSON string) (domain.AxisOrder, error) {
	var projInfo Info
	if err := json.Unmarshal([]byte(projJSON), &projInfo); err != nil {
		return -1, fmt.Errorf("failed to parse %s output: %w", projInfoTool, err)
	}
	if len(projInfo.CoordinateSystem.Axis) < 1 {
		return -1, fmt.Errorf("invalid %s output: axis not found", projInfoTool)
	}

	// east/north == XY, north/east == YX.
	if projInfo.CoordinateSystem.Axis[0].Direction == "north" {
		return domain.AxisOrderYX, nil
	}

	return domain.AxisOrderXY, nil
}

func ExecProjInfo(epsgCode string) (string, error) {
	_, err := execLookPath(projInfoTool)
	if err != nil {
		return "", fmt.Errorf("%s command not found in PATH: %w", projInfoTool, err)
	}

	// Run 'projinfo' and return output in PROJJSON format (https://proj.org/en/stable/specifications/projjson.html)
	cmd := execCommand(projInfoTool, epsgCode, "-o", "projjson", "--single-line", "-q")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute %s command: %w", projInfoTool, err)
	}
	outputJSON := strings.TrimSpace(string(output))

	return outputJSON, nil
}
