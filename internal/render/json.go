package render

import (
	"encoding/json"

	"roastgit/internal/model"
)

// JSON renders the report as pretty JSON.
func JSON(report model.Report) (string, error) {
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
