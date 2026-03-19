// Package models defines Go structs for APsystems OpenAPI responses.
package models

import "encoding/json"

// APIResponse is the generic wrapper returned by every APsystems endpoint.
// The Data field is a json.RawMessage so each handler can decode it into the
// concrete type it expects.
type APIResponse struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

// SystemDetails represents the response from GET /user/api/v2/systems/details/{sid}
type SystemDetails struct {
	SID               string   `json:"sid"`
	CreateDate        string   `json:"create_date"`
	Capacity          string   `json:"capacity"`
	Type              int      `json:"type"`
	Timezone          string   `json:"timezone"`
	ECU               []string `json:"ecu"`
	Light             int      `json:"light"`
	AuthorizationCode string   `json:"authorization_code"`
}

// EnergySummary represents the response from summary endpoints
// (system, ECU) that return today/month/year/lifetime totals.
type EnergySummary struct {
	Today    string `json:"today"`
	Month    string `json:"month"`
	Year     string `json:"year"`
	Lifetime string `json:"lifetime"`
}

// EnergyList is a simple list of energy values returned by period queries
// when the energy_level is hourly/daily/monthly/yearly.
type EnergyList []string

// ECUEnergyMinutely is the object returned when energy_level == "minutely"
// for ECU-level queries.
type ECUEnergyMinutely struct {
	Time   []string `json:"time"`
	Energy []string `json:"energy"`
	Power  []string `json:"power"`
	Today  string   `json:"today"`
}

// InverterInfo groups an ECU with its connected inverters,
// returned by GET /user/api/v2/systems/inverters/{sid}
type InverterInfo struct {
	EID      string     `json:"eid"`
	Type     int        `json:"type"`
	Timezone string     `json:"timezone"`
	Inverter []Inverter `json:"inverter"`
	Model    string     `json:"model,omitempty"`
	Capacity string     `json:"capacity,omitempty"`
}

// Inverter represents a single micro-inverter inside an InverterInfo.
type Inverter struct {
	UID  string `json:"uid"`
	Type string `json:"type"`
}

// InverterSummary holds per-channel energy data for a single inverter.
type InverterSummary struct {
	D1 string `json:"d1,omitempty"`
	M1 string `json:"m1,omitempty"`
	Y1 string `json:"y1,omitempty"`
	T1 string `json:"t1,omitempty"`
	D2 string `json:"d2,omitempty"`
	M2 string `json:"m2,omitempty"`
	Y2 string `json:"y2,omitempty"`
	T2 string `json:"t2,omitempty"`
	D3 string `json:"d3,omitempty"`
	M3 string `json:"m3,omitempty"`
	Y3 string `json:"y3,omitempty"`
	T3 string `json:"t3,omitempty"`
	D4 string `json:"d4,omitempty"`
	M4 string `json:"m4,omitempty"`
	Y4 string `json:"y4,omitempty"`
	T4 string `json:"t4,omitempty"`
}

// InverterEnergyPeriod is the per-channel energy returned for
// hourly/daily/monthly/yearly inverter queries.
type InverterEnergyPeriod struct {
	E1 []string `json:"e1,omitempty"`
	E2 []string `json:"e2,omitempty"`
	E3 []string `json:"e3,omitempty"`
	E4 []string `json:"e4,omitempty"`
}

// InverterMinutelyData holds the detailed telemetry for minutely queries.
type InverterMinutelyData struct {
	T    []string `json:"t"`
	DcP1 []string `json:"dc_p1,omitempty"`
	DcP2 []string `json:"dc_p2,omitempty"`
	DcP3 []string `json:"dc_p3,omitempty"`
	DcP4 []string `json:"dc_p4,omitempty"`
	DcI1 []string `json:"dc_i1,omitempty"`
	DcI2 []string `json:"dc_i2,omitempty"`
	DcI3 []string `json:"dc_i3,omitempty"`
	DcI4 []string `json:"dc_i4,omitempty"`
	DcV1 []string `json:"dc_v1,omitempty"`
	DcV2 []string `json:"dc_v2,omitempty"`
	DcV3 []string `json:"dc_v3,omitempty"`
	DcV4 []string `json:"dc_v4,omitempty"`
	DcE1 []string `json:"dc_e1,omitempty"`
	DcE2 []string `json:"dc_e2,omitempty"`
	DcE3 []string `json:"dc_e3,omitempty"`
	DcE4 []string `json:"dc_e4,omitempty"`
	AcV1 []string `json:"ac_v1,omitempty"`
	AcV2 []string `json:"ac_v2,omitempty"`
	AcV3 []string `json:"ac_v3,omitempty"`
	AcT  []string `json:"ac_t,omitempty"`
	AcP  []string `json:"ac_p,omitempty"`
	AcF  []string `json:"ac_f,omitempty"`
}

// MeterSummary holds consumed/exported/imported/produced totals.
type MeterSummary struct {
	Today    MeterEnergy `json:"today"`
	Month    MeterEnergy `json:"month"`
	Year     MeterEnergy `json:"year"`
	Lifetime MeterEnergy `json:"lifetime"`
}

// MeterEnergy holds the four energy flow values.
type MeterEnergy struct {
	Consumed string `json:"consumed"`
	Exported string `json:"exported"`
	Imported string `json:"imported"`
	Produced string `json:"produced"`
}

// StorageLatest holds the latest power snapshot for a storage ECU.
type StorageLatest struct {
	Mode      string `json:"mode"`
	SOC       string `json:"soc"`
	Time      string `json:"time"`
	Discharge string `json:"discharge"`
	Charge    string `json:"charge"`
	Produced  string `json:"produced"`
	Consumed  string `json:"consumed"`
	Exported  string `json:"exported"`
	Imported  string `json:"imported"`
}

// StorageSummary holds today/month/year/lifetime maps of storage energy.
type StorageSummary struct {
	Today    StorageEnergy `json:"today"`
	Month    StorageEnergy `json:"month"`
	Year     StorageEnergy `json:"year"`
	Lifetime StorageEnergy `json:"lifetime"`
}

// StorageEnergy holds the six storage-specific energy flow values.
type StorageEnergy struct {
	Discharge string `json:"discharge"`
	Charge    string `json:"charge"`
	Produced  string `json:"produced"`
	Consumed  string `json:"consumed"`
	Exported  string `json:"exported"`
	Imported  string `json:"imported"`
}

// APIErrorCodes maps known APsystems response codes to human descriptions.
var APIErrorCodes = map[int]string{
	0:    "Success",
	1000: "Data exception",
	1001: "No data",
	2000: "Application account exception",
	2001: "Invalid application account",
	2002: "Application account is not authorized",
	2003: "Application account authorization expired",
	2004: "Application account has no permission",
	2005: "Access limit of the application account exceeded",
	3000: "Access token exception",
	3001: "Missing access token",
	3002: "Unable to verify access token",
	3003: "Access token timeout",
	3004: "Refresh token timeout",
	4000: "Request parameter exception",
	4001: "Invalid request parameter",
	5000: "Internal server exception",
	6000: "Communication exception",
	7000: "Server access restriction exception",
	7001: "Server access limit exceeded",
	7002: "Too many requests, please request later",
	7003: "System is busy, please request later",
}
