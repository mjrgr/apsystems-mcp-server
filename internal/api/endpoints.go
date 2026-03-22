package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mjrgr/apsystems-mcp-server/internal/models"
)

// ---------- System Details ----------

// GetSystemDetails returns details for a particular system.
func (c *Client) GetSystemDetails(ctx context.Context, sid string) (*models.SystemDetails, error) {
	var out models.SystemDetails
	_, err := c.Do(ctx, "GET",
		fmt.Sprintf("/user/api/v2/systems/details/%s", sid),
		nil, &out,
	)
	return &out, err
}

// GetSystemInverters returns all ECUs and their inverters for a system.
func (c *Client) GetSystemInverters(ctx context.Context, sid string) ([]models.InverterInfo, error) {
	var out []models.InverterInfo
	_, err := c.Do(ctx, "GET",
		fmt.Sprintf("/user/api/v2/systems/inverters/%s", sid),
		nil, &out,
	)
	return out, err
}

// GetSystemMeters returns all meter IDs for a system.
func (c *Client) GetSystemMeters(ctx context.Context, sid string) ([]string, error) {
	var out []string
	_, err := c.Do(ctx, "GET",
		fmt.Sprintf("/user/api/v2/systems/meters/%s", sid),
		nil, &out,
	)
	return out, err
}

// ---------- System-level Data ----------

// GetSystemSummary returns the accumulative energy summary for a system.
func (c *Client) GetSystemSummary(ctx context.Context, sid string) (*models.EnergySummary, error) {
	var out models.EnergySummary
	_, err := c.Do(ctx, "GET",
		fmt.Sprintf("/user/api/v2/systems/summary/%s", sid),
		nil, &out,
	)
	return &out, err
}

// GetSystemEnergy returns period energy data for a system.
// energyLevel: hourly|daily|monthly|yearly
// dateRange: format depends on energyLevel (may be empty for yearly).
func (c *Client) GetSystemEnergy(ctx context.Context, sid, energyLevel, dateRange string) (json.RawMessage, error) {
	q := map[string]string{"energy_level": energyLevel}
	if dateRange != "" {
		q["date_range"] = dateRange
	}
	raw, err := c.Do(ctx, "GET",
		fmt.Sprintf("/user/api/v2/systems/energy/%s", sid),
		q, nil,
	)
	return raw, err
}

// ---------- ECU-level Data ----------

// GetECUSummary returns the accumulative energy summary for an ECU.
func (c *Client) GetECUSummary(ctx context.Context, sid, eid string) (*models.EnergySummary, error) {
	var out models.EnergySummary
	_, err := c.Do(ctx, "GET",
		fmt.Sprintf("/user/api/v2/systems/%s/devices/ecu/summary/%s", sid, eid),
		nil, &out,
	)
	return &out, err
}

// GetECUEnergy returns period energy data for an ECU.
// energyLevel: minutely|hourly|daily|monthly|yearly
func (c *Client) GetECUEnergy(ctx context.Context, sid, eid, energyLevel, dateRange string) (json.RawMessage, error) {
	q := map[string]string{"energy_level": energyLevel}
	if dateRange != "" {
		q["date_range"] = dateRange
	}
	raw, err := c.Do(ctx, "GET",
		fmt.Sprintf("/user/api/v2/systems/%s/devices/ecu/energy/%s", sid, eid),
		q, nil,
	)
	return raw, err
}

// ---------- Meter-level Data ----------

// GetMeterSummary returns the accumulative energy summary for a meter.
func (c *Client) GetMeterSummary(ctx context.Context, sid, eid string) (*models.MeterSummary, error) {
	var out models.MeterSummary
	_, err := c.Do(ctx, "GET",
		fmt.Sprintf("/user/api/v2/systems/%s/devices/meter/summary/%s", sid, eid),
		nil, &out,
	)
	return &out, err
}

// GetMeterPeriod returns period energy data for a meter.
func (c *Client) GetMeterPeriod(ctx context.Context, sid, eid, energyLevel, dateRange string) (json.RawMessage, error) {
	q := map[string]string{"energy_level": energyLevel}
	if dateRange != "" {
		q["date_range"] = dateRange
	}
	raw, err := c.Do(ctx, "GET",
		fmt.Sprintf("/user/api/v2/systems/%s/devices/meter/period/%s", sid, eid),
		q, nil,
	)
	return raw, err
}

// ---------- Inverter-level Data ----------

// GetInverterSummary returns per-channel energy for a single inverter.
func (c *Client) GetInverterSummary(ctx context.Context, sid, uid string) (*models.InverterSummary, error) {
	var out models.InverterSummary
	_, err := c.Do(ctx, "GET",
		fmt.Sprintf("/user/api/v2/systems/%s/devices/inverter/summary/%s", sid, uid),
		nil, &out,
	)
	return &out, err
}

// GetInverterEnergy returns period energy for a single inverter.
func (c *Client) GetInverterEnergy(ctx context.Context, sid, uid, energyLevel, dateRange string) (json.RawMessage, error) {
	q := map[string]string{"energy_level": energyLevel}
	if dateRange != "" {
		q["date_range"] = dateRange
	}
	raw, err := c.Do(ctx, "GET",
		fmt.Sprintf("/user/api/v2/systems/%s/devices/inverter/energy/%s", sid, uid),
		q, nil,
	)
	return raw, err
}

// GetInverterBatchEnergy returns energy/power for all inverters under an ECU in a day.
func (c *Client) GetInverterBatchEnergy(ctx context.Context, sid, eid, energyLevel, dateRange string) (json.RawMessage, error) {
	q := map[string]string{
		"energy_level": energyLevel,
		"date_range":   dateRange,
	}
	raw, err := c.Do(ctx, "GET",
		fmt.Sprintf("/user/api/v2/systems/%s/devices/inverter/batch/energy/%s", sid, eid),
		q, nil,
	)
	return raw, err
}

// ---------- Storage-level Data ----------

// GetStorageLatest returns the latest status of a storage ECU.
func (c *Client) GetStorageLatest(ctx context.Context, sid, eid string) (*models.StorageLatest, error) {
	var out models.StorageLatest
	_, err := c.Do(ctx, "GET",
		fmt.Sprintf("/installer/api/v2/systems/%s/devices/storage/latest/%s", sid, eid),
		nil, &out,
	)
	return &out, err
}

// GetStorageSummary returns the accumulative energy summary for a storage ECU.
func (c *Client) GetStorageSummary(ctx context.Context, sid, eid string) (*models.StorageSummary, error) {
	var out models.StorageSummary
	_, err := c.Do(ctx, "GET",
		fmt.Sprintf("/installer/api/v2/systems/%s/devices/storage/summary/%s", sid, eid),
		nil, &out,
	)
	return &out, err
}

// GetStoragePeriod returns period energy data for a storage ECU.
func (c *Client) GetStoragePeriod(ctx context.Context, sid, eid, energyLevel, dateRange string) (json.RawMessage, error) {
	q := map[string]string{"energy_level": energyLevel}
	if dateRange != "" {
		q["date_range"] = dateRange
	}
	raw, err := c.Do(ctx, "GET",
		fmt.Sprintf("/installer/api/v2/systems/%s/devices/storage/period/%s", sid, eid),
		q, nil,
	)
	return raw, err
}
