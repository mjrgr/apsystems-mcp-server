# API Reference

All tools communicate via the MCP protocol (JSON-RPC 2.0 over stdio). Each tool accepts structured input arguments and returns JSON content.

## Common Patterns

- **System ID (`sid`)** — The unique identifier for an APsystems solar system (e.g. `"AZ12649A3DFF"`)
- **ECU ID (`eid`)** — The Energy Communication Unit identifier (e.g. `"203000001234"`)
- **Inverter UID (`uid`)** — A micro-inverter identifier (e.g. `"902000001234"`)
- **Energy Level** — Time granularity: `"minutely"`, `"hourly"`, `"daily"`, `"monthly"`, `"yearly"`
- **Date Range** — Date filter whose format depends on the energy level:
  - `minutely` / `hourly` → `"yyyy-MM-dd"` (e.g. `"2024-06-15"`)
  - `daily` → `"yyyy-MM"` (e.g. `"2024-06"`)
  - `monthly` → `"yyyy"` (e.g. `"2024"`)
  - `yearly` → omit this field
- **Energy units** — All energy values are strings in kWh. Power values are in W.

---

## System Details API

### `get_system_details`

Returns system metadata including capacity, type, ECU list, and status.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |

**Returns:** `sid`, `create_date`, `capacity` (kW), `type` (1=PV, 2=Storage, 3=Both), `timezone`, `ecu` (list), `light` (1=Green, 2=Yellow, 3=Red, 4=Grey), `authorization_code`

#### Example
**Request:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_system_details",
    "arguments": { "sid": "AZ12649A3DFF" }
  }
}
```
**Response:**
```json
{
  "content": [
    {
      "type": "json",
      "data": {
        "sid": "AZ12649A3DFF",
        "create_date": "2023-01-15",
        "capacity": "7.2",
        "type": 1,
        "timezone": "America/Los_Angeles",
        "ecu": ["203000001234"],
        "light": 1,
        "authorization_code": "..."
      }
    }
  ],
  "isError": false
}
```

---

### `get_system_inverters`

Lists all ECUs and their connected micro-inverters.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |

**Returns:** Array of ECUs, each with `eid`, `type` (0=ECU, 1=with meter, 2=with storage), `timezone`, `inverter` (list of `{uid, type}`)

---

### `get_system_meters`

Lists all meter IDs for a system.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |

**Returns:** Array of meter ID strings

---

## System-level Data API

### `get_system_summary`

Returns cumulative energy totals.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |

**Returns:** `today`, `month`, `year`, `lifetime` (all kWh strings)

#### Example
**Request:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_system_summary",
    "arguments": { "sid": "AZ12649A3DFF" }
  }
}
```
**Response:**
```json
{
  "content": [
    {
      "type": "json",
      "data": {
        "today": "12.28",
        "month": "320.1",
        "year": "1200.5",
        "lifetime": "4500.2"
      }
    }
  ],
  "isError": false
}
```

---

### `get_system_energy`

Returns energy data at a specified time granularity.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |
| `energy_level` | string | Yes | `"hourly"`, `"daily"`, `"monthly"`, `"yearly"` |
| `date_range` | string | No | Date filter (format depends on energy_level) |

**Returns:** Array of energy value strings. Length: 24 (hourly), days-in-month (daily), 12 (monthly), years-since-install (yearly).

---

## ECU-level Data API

### `get_ecu_summary`

Returns cumulative energy for a specific ECU.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |
| `eid` | string | Yes | ECU ID |

**Returns:** `today`, `month`, `year`, `lifetime` (all kWh strings)

#### Example
**Request:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_ecu_summary",
    "arguments": { "sid": "AZ12649A3DFF", "eid": "203000001234" }
  }
}
```
**Response:**
```json
{
  "content": [
    {
      "type": "json",
      "data": {
        "today": "5.12",
        "month": "140.3",
        "year": "600.7",
        "lifetime": "2200.9"
      }
    }
  ],
  "isError": false
}
```

---

### `get_ecu_energy`

Returns ECU energy data at a specified granularity. Supports `minutely` for power telemetry.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |
| `eid` | string | Yes | ECU ID |
| `energy_level` | string | Yes | `"minutely"`, `"hourly"`, `"daily"`, `"monthly"`, `"yearly"` |
| `date_range` | string | No | Date filter |

**Returns:**
- For `hourly`/`daily`/`monthly`/`yearly`: Array of kWh strings
- For `minutely`: Object with `time` (list), `energy` (list, kWh), `power` (list, W), `today` (kWh)

---

## Meter-level Data API

### `get_meter_summary`

Returns consumed/exported/imported/produced energy totals for a meter.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |
| `eid` | string | Yes | Meter ECU ID |

**Returns:** Object with `today`, `month`, `year`, `lifetime` — each containing `consumed`, `exported`, `imported`, `produced` (kWh strings)

---

### `get_meter_energy`

Returns meter energy data at a specified granularity.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |
| `eid` | string | Yes | Meter ECU ID |
| `energy_level` | string | Yes | `"minutely"`, `"hourly"`, `"daily"`, `"monthly"`, `"yearly"` |
| `date_range` | string | No | Date filter |

**Returns:**
- For `hourly`/`daily`/`monthly`/`yearly`: Object with `time`, `produced`, `consumed`, `imported`, `exported` arrays
- For `minutely`: Object with `time`, `power` (map), `energy` (map), `today` (map)

---

## Inverter-level Data API

### `get_inverter_summary`

Returns per-channel energy summary for a single inverter (up to 4 DC channels).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |
| `uid` | string | Yes | Inverter UID |

**Returns:** Object with fields `d1`..`d4` (today), `m1`..`m4` (month), `y1`..`y4` (year), `t1`..`t4` (lifetime). All kWh strings. Only populated channels are present.

#### Example
**Request:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_inverter_summary",
    "arguments": { "sid": "AZ12649A3DFF", "uid": "902000001234" }
  }
}
```
**Response:**
```json
{
  "content": [
    {
      "type": "json",
      "data": {
        "d1": "1.23",
        "d2": "1.20",
        "m1": "30.1",
        "m2": "29.8",
        "y1": "120.5",
        "y2": "119.7",
        "t1": "400.2",
        "t2": "398.9"
      }
    }
  ],
  "isError": false
}
```

---

### `get_inverter_energy`

Returns per-channel energy data at a specified granularity.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |
| `uid` | string | Yes | Inverter UID |
| `energy_level` | string | Yes | `"minutely"`, `"hourly"`, `"daily"`, `"monthly"`, `"yearly"` |
| `date_range` | string | No | Date filter |

**Returns:**
- For `hourly`/`daily`/`monthly`/`yearly`: Object with `e1`..`e4` (per-channel energy arrays)
- For `minutely`: Object with `t` (times), `dc_p1`..`dc_p4` (DC power), `dc_i1`..`dc_i4` (DC current), `dc_v1`..`dc_v4` (DC voltage), `dc_e1`..`dc_e4` (DC energy), `ac_v1`..`ac_v3`, `ac_t`, `ac_p`, `ac_f`

---

### `get_inverter_batch_energy`

Returns energy or power for ALL inverters under an ECU for a single day.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |
| `eid` | string | Yes | ECU ID |
| `energy_level` | string | Yes | `"power"` or `"energy"` |
| `date_range` | string | Yes | Date in `"yyyy-MM-dd"` format |

**Returns:**
- For `energy`: Object with `energy` array (format: `"uid-channel-kWh"`)
- For `power`: Object with `time` array and `power` map (`{uid-channel: [values]}`)

---

## Storage-level Data API

### `get_storage_latest`

Returns the latest power snapshot for a storage ECU.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |
| `eid` | string | Yes | Storage ECU ID |

**Returns:** `mode`, `soc` (%), `time`, `discharge`, `charge`, `produced`, `consumed`, `exported`, `imported`

---

### `get_storage_summary`

Returns cumulative storage energy totals.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |
| `eid` | string | Yes | Storage ECU ID |

**Returns:** Object with `today`, `month`, `year`, `lifetime` — each containing `discharge`, `charge`, `produced`, `consumed`, `exported`, `imported` (kWh strings)

---

### `get_storage_energy`

Returns storage energy data at a specified granularity.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sid` | string | Yes | System ID |
| `eid` | string | Yes | Storage ECU ID |
| `energy_level` | string | Yes | `"minutely"`, `"hourly"`, `"daily"`, `"monthly"`, `"yearly"` |
| `date_range` | string | No | Date filter |

**Returns:**
- For `hourly`/`daily`/`monthly`/`yearly`: Object with `time`, `discharge`, `charge`, `produced`, `consumed`, `exported`, `imported` arrays
- For `minutely`: Object with `time`, `power` (map), `energy` (map), `today` (map)

---

## Error Responses

Tool errors are returned with `isError: true` and a text description:

```json
{
  "content": [{"type": "text", "text": "Error: missing required parameters: sid"}],
  "isError": true
}
```

APsystems API errors include the numeric code and description:

```json
{
  "content": [{"type": "text", "text": "Error: APsystems API error 2001: Invalid application account"}],
  "isError": true
}
```

See the [APsystems error code reference](../README.md#api-error-codes) for the full list.
