package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	db *pgxpool.Pool
}

type Aircraft struct {
	ID                                 int      `json:"id"`
	ICAOCode                          string   `json:"icao_code"`
	FAADesignator                     string   `json:"faa_designator"`
	Manufacturer                      string   `json:"manufacturer"`
	ModelFAA                          string   `json:"model_faa"`
	ModelBADA                         string   `json:"model_bada"`
	PhysicalClassEngine               string   `json:"physical_class_engine"`
	NumEngines                        *int     `json:"num_engines"`
	AAC                               string   `json:"aac"`
	AACMinimum                        string   `json:"aac_minimum"`
	AACMaximum                        string   `json:"aac_maximum"`
	ADG                               string   `json:"adg"`
	TDG                               string   `json:"tdg"`
	ApproachSpeedKnot                 *int     `json:"approach_speed_knot"`
	ApproachSpeedMinimumKnot          *int     `json:"approach_speed_minimum_knot"`
	ApproachSpeedMaximumKnot          *int     `json:"approach_speed_maximum_knot"`
	WingspanFtWithoutWingletsSharklets *float64 `json:"wingspan_ft_without_winglets_sharklets"`
	WingspanFtWithWingletsSharklets   *float64 `json:"wingspan_ft_with_winglets_sharklets"`
	LengthFt                          *float64 `json:"length_ft"`
	TailHeightAtOEWFt                 *float64 `json:"tail_height_at_oew_ft"`
	WheelbaseFt                       *float64 `json:"wheelbase_ft"`
	CockpitToMainGearFt               *float64 `json:"cockpit_to_main_gear_ft"`
	MainGearWidthFt                   *float64 `json:"main_gear_width_ft"`
	MTOWLB                            *int     `json:"mtow_lb"`
	MALWLB                            *int     `json:"malw_lb"`
	MainGearConfig                    string   `json:"main_gear_config"`
	ICAOWTC                           string   `json:"icao_wtc"`
	ParkingAreaFt2                    *float64 `json:"parking_area_ft2"`
	Class                             string   `json:"class"`
	FAAWeight                         string   `json:"faa_weight"`
	CWT                               string   `json:"cwt"`
	OneHalfWakeCategory               string   `json:"one_half_wake_category"`
	TwoWakeCategoryAppxA              string   `json:"two_wake_category_appx_a"`
	TwoWakeCategoryAppxB              string   `json:"two_wake_category_appx_b"`
	RotorDiameterFt                   *float64 `json:"rotor_diameter_ft"`
	SRS                               string   `json:"srs"`
	LAHSO                             string   `json:"lahso"`
	FAARegistry                       string   `json:"faa_registry"`
	RegistrationCount                 *int     `json:"registration_count"`
	TMFSOperationsFY24                *int     `json:"tmfs_operations_fy24"`
	Remarks                           string   `json:"remarks"`
	LastUpdate                        string   `json:"last_update"`
	CreatedAt                         time.Time `json:"created_at"`
	UpdatedAt                         time.Time `json:"updated_at"`
}

type SearchResponse struct {
	Aircraft []Aircraft `json:"aircraft"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	Limit    int        `json:"limit"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func New(db *pgxpool.Pool) *Handlers {
	return &Handlers{db: db}
}

// SearchAircraft handles GET /api/aircraft/search
func (h *Handlers) SearchAircraft(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Parse query parameters
	query := r.URL.Query()
	search := query.Get("q")
	pageStr := query.Get("page")
	limitStr := query.Get("limit")

	// Default pagination
	page := 1
	limit := 50

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	// Build search query
	var whereClause string
	var args []interface{}
	argCount := 0

	if search != "" {
		searchTerm := "%" + strings.ToUpper(search) + "%"
		whereClause = `WHERE 
			UPPER(icao_code) LIKE $1 OR 
			UPPER(faa_designator) LIKE $1 OR 
			UPPER(manufacturer) LIKE $1 OR 
			UPPER(model_faa) LIKE $1`
		args = append(args, searchTerm)
		argCount = 1
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM aircraft_data %s", whereClause)
	var total int
	err := h.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get aircraft records
	dataQuery := fmt.Sprintf(`
		SELECT id, icao_code, faa_designator, manufacturer, model_faa, model_bada,
			   physical_class_engine, num_engines, aac, aac_minimum, aac_maximum,
			   adg, tdg, approach_speed_knot, approach_speed_minimum_knot, approach_speed_maximum_knot,
			   wingspan_ft_without_winglets_sharklets, wingspan_ft_with_winglets_sharklets,
			   length_ft, tail_height_at_oew_ft, wheelbase_ft, cockpit_to_main_gear_ft,
			   main_gear_width_ft, mtow_lb, malw_lb, main_gear_config, icao_wtc,
			   parking_area_ft2, class, faa_weight, cwt, one_half_wake_category,
			   two_wake_category_appx_a, two_wake_category_appx_b, rotor_diameter_ft,
			   srs, lahso, faa_registry, registration_count, tmfs_operations_fy24,
			   remarks, last_update, created_at, updated_at
		FROM aircraft_data %s 
		ORDER BY manufacturer, model_faa 
		LIMIT $%d OFFSET $%d`, whereClause, argCount+1, argCount+2)

	args = append(args, limit, offset)

	rows, err := h.db.Query(ctx, dataQuery, args...)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var aircraft []Aircraft
	for rows.Next() {
		var a Aircraft
		err := rows.Scan(
			&a.ID, &a.ICAOCode, &a.FAADesignator, &a.Manufacturer, &a.ModelFAA, &a.ModelBADA,
			&a.PhysicalClassEngine, &a.NumEngines, &a.AAC, &a.AACMinimum, &a.AACMaximum,
			&a.ADG, &a.TDG, &a.ApproachSpeedKnot, &a.ApproachSpeedMinimumKnot, &a.ApproachSpeedMaximumKnot,
			&a.WingspanFtWithoutWingletsSharklets, &a.WingspanFtWithWingletsSharklets,
			&a.LengthFt, &a.TailHeightAtOEWFt, &a.WheelbaseFt, &a.CockpitToMainGearFt,
			&a.MainGearWidthFt, &a.MTOWLB, &a.MALWLB, &a.MainGearConfig, &a.ICAOWTC,
			&a.ParkingAreaFt2, &a.Class, &a.FAAWeight, &a.CWT, &a.OneHalfWakeCategory,
			&a.TwoWakeCategoryAppxA, &a.TwoWakeCategoryAppxB, &a.RotorDiameterFt,
			&a.SRS, &a.LAHSO, &a.FAARegistry, &a.RegistrationCount, &a.TMFSOperationsFY24,
			&a.Remarks, &a.LastUpdate, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		aircraft = append(aircraft, a)
	}

	response := SearchResponse{
		Aircraft: aircraft,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAircraft handles GET /api/aircraft/{id}
func (h *Handlers) GetAircraft(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Extract ID from URL path
	path := r.URL.Path
	idStr := strings.TrimPrefix(path, "/api/aircraft/")
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid aircraft ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, icao_code, faa_designator, manufacturer, model_faa, model_bada,
			   physical_class_engine, num_engines, aac, aac_minimum, aac_maximum,
			   adg, tdg, approach_speed_knot, approach_speed_minimum_knot, approach_speed_maximum_knot,
			   wingspan_ft_without_winglets_sharklets, wingspan_ft_with_winglets_sharklets,
			   length_ft, tail_height_at_oew_ft, wheelbase_ft, cockpit_to_main_gear_ft,
			   main_gear_width_ft, mtow_lb, malw_lb, main_gear_config, icao_wtc,
			   parking_area_ft2, class, faa_weight, cwt, one_half_wake_category,
			   two_wake_category_appx_a, two_wake_category_appx_b, rotor_diameter_ft,
			   srs, lahso, faa_registry, registration_count, tmfs_operations_fy24,
			   remarks, last_update, created_at, updated_at
		FROM aircraft_data 
		WHERE id = $1`

	var aircraft Aircraft
	err = h.db.QueryRow(ctx, query, id).Scan(
		&aircraft.ID, &aircraft.ICAOCode, &aircraft.FAADesignator, &aircraft.Manufacturer, &aircraft.ModelFAA, &aircraft.ModelBADA,
		&aircraft.PhysicalClassEngine, &aircraft.NumEngines, &aircraft.AAC, &aircraft.AACMinimum, &aircraft.AACMaximum,
		&aircraft.ADG, &aircraft.TDG, &aircraft.ApproachSpeedKnot, &aircraft.ApproachSpeedMinimumKnot, &aircraft.ApproachSpeedMaximumKnot,
		&aircraft.WingspanFtWithoutWingletsSharklets, &aircraft.WingspanFtWithWingletsSharklets,
		&aircraft.LengthFt, &aircraft.TailHeightAtOEWFt, &aircraft.WheelbaseFt, &aircraft.CockpitToMainGearFt,
		&aircraft.MainGearWidthFt, &aircraft.MTOWLB, &aircraft.MALWLB, &aircraft.MainGearConfig, &aircraft.ICAOWTC,
		&aircraft.ParkingAreaFt2, &aircraft.Class, &aircraft.FAAWeight, &aircraft.CWT, &aircraft.OneHalfWakeCategory,
		&aircraft.TwoWakeCategoryAppxA, &aircraft.TwoWakeCategoryAppxB, &aircraft.RotorDiameterFt,
		&aircraft.SRS, &aircraft.LAHSO, &aircraft.FAARegistry, &aircraft.RegistrationCount, &aircraft.TMFSOperationsFY24,
		&aircraft.Remarks, &aircraft.LastUpdate, &aircraft.CreatedAt, &aircraft.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			http.Error(w, "Aircraft not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(aircraft)
}

// HealthCheck handles GET /api/health
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Test database connection
	err := h.db.Ping(ctx)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Database unavailable"})
		return
	}

	// Get record count
	var count int
	err = h.db.QueryRow(ctx, "SELECT COUNT(*) FROM aircraft_data").Scan(&count)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Database query failed"})
		return
	}

	response := map[string]interface{}{
		"status":          "healthy",
		"database":        "connected",
		"aircraft_count":  count,
		"timestamp":       time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}