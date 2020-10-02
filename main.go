package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"golang.org/x/net/publicsuffix"
)

var baseURL = "https://api.onepeloton.com"

type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email"`
	Password        string `json:"password"`
	WithPubSub      bool   `json:"with_pubsub"`
}

type ErrorResponse struct {
	Status    int    `json:"status"`
	ErrorCode int    `json:"error_code"`
	SubCode   int    `json:"subcode,omitempty"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
}

type WorkoutsResponse struct {
	Data []struct {
		CreatedAt                  int     `json:"created_at"`
		DeviceType                 string  `json:"device_type"`
		EndTime                    int     `json:"end_time"`
		FitbitID                   string  `json:"fitbit_id,omitempty"`
		FitnessDiscipline          string  `json:"fitness_discipline"`
		HasPedalingMetrics         bool    `json:"has_pedaling_metrics"`
		HasLeaderboardMetrics      bool    `json:"has_leaderboard_metrics"`
		ID                         string  `json:"id"`
		IsTotalWorkPersonalRecord  bool    `json:"is_total_work_personal_record"`
		MetricsType                string  `json:"metrics_type"`
		Name                       string  `json:"name"`
		PelotonID                  string  `json:"peloton_id"`
		Platform                   string  `json:"platform"`
		StartTime                  int     `json:"start_time"`
		StravaID                   string  `json:"strava_id"`
		Status                     string  `json:"status"`
		Timezone                   string  `json:"timezone"`
		Title                      string  `json:"title,omitempty"`
		TotalWork                  float64 `json:"total_work"`
		UserID                     string  `json:"user_id"`
		WorkoutType                string  `json:"workout_type"`
		TotalVideoWatchTimeSeconds int     `json:"total_video_watch_time_seconds"`
		TotalVideoBufferingSeconds int     `json:"total_video_buffering_seconds"`
		Ride                       struct {
			HasClosedCaptions            bool     `json:"has_closed_captions"`
			ContentProvider              string   `json:"content_provider"`
			ContentFormat                string   `json:"content_format"`
			Description                  string   `json:"description"`
			DifficultyRatingAvg          float64  `json:"difficulty_rating_avg"`
			DifficultyRatingCount        int      `json:"difficulty_rating_count"`
			DifficultyLevel              string   `json:"difficulty_level"`
			Duration                     int      `json:"duration"`
			ExtraImages                  []string `json:"extra_images,omitempty"`
			FitnessDiscipline            string   `json:"fitness_discipline"`
			FitnessDisciplineDisplayName string   `json:"fitness_discipline_display_name"`
			HasPedalingMetrics           bool     `json:"has_pedaling_metrics"`
			HomePelotonID                string   `json:"home_peloton_id"`
			ID                           string   `json:"id"`
			ImageURL                     string   `json:"image_url"`
			InstructorID                 string   `json:"instructor_id"`
			IsArchived                   bool     `json:"is_archived"`
			IsClosedCaptionShown         bool     `json:"is_closed_caption_shown"`
			IsExplicit                   bool     `json:"is_explicit"`
			IsLiveInStudioOnly           bool     `json:"is_live_in_studio_only"`
			Language                     string   `json:"language"`
			Length                       int      `json:"length"`
			LiveStreamID                 string   `json:"live_stream_id"`
			LiveStreamURL                string   `json:"live_stream_url,omitempty"`
			Location                     string   `json:"location"`
			Metrics                      []string `json:"metrics"`
			OriginalAirTime              int      `json:"original_air_time"`
			OverallRatingAvg             float64  `json:"overall_rating_avg"`
			OverallRatingCount           int      `json:"overall_rating_count"`
			PedalingStartOffset          int      `json:"pedaling_start_offset"`
			PedalingEndOffset            int      `json:"pedaling_end_offset"`
			PedalingDuration             int      `json:"pedaling_duration"`
			Rating                       int      `json:"rating"`
			RideTypeID                   string   `json:"ride_type_id"`
			RideTypeIds                  []string `json:"ride_type_ids"`
			SampleVodStreamURL           string   `json:"sample_vod_stream_url,omitempty"`
			ScheduledStartTime           int      `json:"scheduled_start_time"`
			SeriesID                     string   `json:"series_id"`
			SoldOut                      bool     `json:"sold_out"`
			StudioPelotonID              string   `json:"studio_peloton_id"`
			Title                        string   `json:"title"`
			TotalRatings                 int      `json:"total_ratings"`
			TotalInProgressWorkouts      int      `json:"total_in_progress_workouts"`
			TotalWorkouts                int      `json:"total_workouts"`
			VodStreamURL                 string   `json:"vod_stream_url"`
			VodStreamID                  string   `json:"vod_stream_id"`
			ClassTypeIds                 []string `json:"class_type_ids"`
			DifficultyEstimate           float64  `json:"difficulty_estimate"`
			OverallEstimate              float64  `json:"overall_estimate"`
		} `json:"ride"`
		Created             int    `json:"created"`
		DeviceTimeCreatedAt int    `json:"device_time_created_at"`
		EffortZones         string `json:"effort_zones,omitempty"`
	} `json:"data"`
	Limit          int            `json:"limit"`
	Page           int            `json:"page"`
	Total          int            `json:"total"`
	Count          int            `json:"count"`
	PageCount      int            `json:"page_count"`
	ShowPrevious   bool           `json:"show_previous"`
	ShowNext       bool           `json:"show_next"`
	SortBy         string         `json:"sort_by"`
	Summary        map[string]int `json:"summary"`
	AggregateStats []string       `json:"aggregate_stats,omitempty"`
}

type MeResponse struct {
	InstructorID       string  `json:"instructor_id,omitempty"`
	FacebookID         string  `json:"facebook_id,omitempty"`
	Email              string  `json:"email"`
	LastName           string  `json:"last_name"`
	Weight             float64 `json:"weight"`
	ContractAgreements []struct {
		ContractType        string `json:"contract_type"`
		ContractID          string `json:"contract_id"`
		ContractCreatedAt   int    `json:"contract_created_at"`
		BikeContractURL     string `json:"bike_contract_url"`
		TreadContractURL    string `json:"tread_contract_url"`
		AgreedAt            int    `json:"agreed_at"`
		ContractDisplayName string `json:"contract_display_name"`
	} `json:"contract_agreements"`
	EstimatedCyclingFtp          int  `json:"estimated_cycling_ftp"`
	ReferralsMade                int  `json:"referrals_made"`
	DefaultMaxHeartRate          int  `json:"default_max_heart_rate"`
	HasActiveDigitalSubscription bool `json:"has_active_digital_subscription"`
	PairedDevices                []struct {
		Name             string `json:"name"`
		PairedDeviceType string `json:"paired_device_type"`
		SerialNumber     string `json:"serial_number"`
	} `json:"paired_devices"`
	MemberGroups                   []string  `json:"member_groups"`
	IsInternalBetaTester           bool      `json:"is_internal_beta_tester"`
	Gender                         string    `json:"gender"`
	DefaultHeartRateZones          []float64 `json:"default_heart_rate_zones"`
	IsProvisional                  bool      `json:"is_provisional"`
	ImageURL                       string    `json:"image_url"`
	CreatedCountry                 string    `json:"created_country"`
	PhoneNumber                    string    `json:"phone_number"`
	LastWorkoutAt                  int       `json:"last_workout_at"`
	Username                       string    `json:"username"`
	IsExternalBetaTester           bool      `json:"is_external_beta_tester"`
	HasActiveDeviceSubscription    bool      `json:"has_active_device_subscription"`
	CyclingFtpSource               string    `json:"cycling_ftp_source"`
	ReferralCode                   string    `json:"referral_code,omitempty"`
	SubscriptionCreditsUsed        int       `json:"subscription_credits_used"`
	TotalNonPedalingMetricWorkouts int       `json:"total_non_pedaling_metric_workouts"`
	Height                         float64   `json:"height"`
	Birthday                       int       `json:"birthday"`
	CustomizedHeartRateZones       []string  `json:"customized_heart_rate_zones,omitempty"`
	SubscriptionCredits            int       `json:"subscription_credits"`
	TotalPedalingMetricWorkouts    int       `json:"total_pedaling_metric_workouts"`
	ExternalMusicAuthList          []struct {
		Provider string `json:"provider"`
		Status   string `json:"status"`
		Email    string `json:"email,omitempty"`
	} `json:"external_music_auth_list"`
	FacebookAccessToken    string `json:"facebook_access_token,omitempty"`
	FirstName              string `json:"first_name"`
	CyclingFtp             int    `json:"cycling_ftp"`
	Name                   string `json:"name"`
	CanCharge              bool   `json:"can_charge"`
	IsStravaAuthenticated  bool   `json:"is_strava_authenticated"`
	ID                     string `json:"id"`
	CyclingWorkoutFtp      int    `json:"cycling_workout_ftp"`
	BlockExplicit          bool   `json:"block_explicit"`
	IsDemo                 bool   `json:"is_demo"`
	CreatedAt              int    `json:"created_at"`
	IsCompleteProfile      bool   `json:"is_complete_profile"`
	TotalWorkouts          int    `json:"total_workouts"`
	TotalPendingFollowers  int    `json:"total_pending_followers"`
	CustomizedMaxHeartRate int    `json:"customized_max_heart_rate"`
	IsProfilePrivate       bool   `json:"is_profile_private"`
	HasSignedWaiver        bool   `json:"has_signed_waiver"`
	QuickHits              struct {
		QuickHitsEnabled bool   `json:"quick_hits_enabled"`
		SpeedShortcuts   string `json:"speed_shortcuts,omitempty"`
		InclineShortcuts string `json:"incline_shortcuts,omitempty"`
	} `json:"quick_hits"`
	ObfuscatedEmail       string `json:"obfuscated_email"`
	TotalFollowers        int    `json:"total_followers"`
	IsFitbitAuthenticated bool   `json:"is_fitbit_authenticated"`
	WorkoutCounts         []struct {
		Name    string `json:"name"`
		Slug    string `json:"slug"`
		Count   int    `json:"count"`
		IconURL string `json:"icon_url"`
	} `json:"workout_counts"`
	CyclingFtpWorkoutID string `json:"cycling_ftp_workout_id,omitempty"`
	TotalFollowing      int    `json:"total_following"`
	V1ReferralsMade     int    `json:"v1_referrals_made"`
	Location            string `json:"location"`
	MiddleInitial       string `json:"middle_initial"`
	HardwareSettings    string `json:"hardware_settings,omitempty"`
}

type Client struct {
	client *http.Client
}

func NewClient() Client {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	return Client{
		client: &http.Client{
			Jar: jar,
		},
	}
}

func (c *Client) Login(username, password string) error {
	l := LoginRequest{
		UsernameOrEmail: username,
		Password:        password,
		WithPubSub:      false,
	}

	login, err := json.Marshal(l)
	if err != nil {
		return err
	}

	res, err := c.client.Post(fmt.Sprintf("%s/auth/login", baseURL), "application/json", bytes.NewBuffer(login))
	if err != nil {
		return err
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	c.client.Jar.SetCookies(u, res.Cookies())
	return nil
}

func (c *Client) Me() (MeResponse, error) {
	var me MeResponse
	res, err := c.client.Get(fmt.Sprintf("%s/api/me", baseURL))
	if err != nil {
		return me, err
	}

	err = json.NewDecoder(res.Body).Decode(&me)
	if err != nil {
		return me, err
	}

	return me, nil
}

func (c *Client) Workouts(userID string) (WorkoutsResponse, error) {
	var workouts WorkoutsResponse
	res, err := c.client.Get(fmt.Sprintf("%s/api/user/%s/workouts", baseURL, userID))
	if err != nil {
		return workouts, err
	}

	err = json.NewDecoder(res.Body).Decode(&workouts)
	if err != nil {
		return workouts, err
	}

	return workouts, nil
}

func (c *Client) WorkoutsCSV(userID string) ([]byte, error) {
	res, err := c.client.Get(fmt.Sprintf("%s/api/user/%s/workout_history_csv", baseURL, userID))

	if err != nil {
		return []byte{}, err
	}

	return ioutil.ReadAll(res.Body)
}

func main() {
	c := NewClient()
	err := c.Login(os.Getenv("USERNAME"), os.Getenv("PASSWORD"))
	if err != nil {
		log.Fatal(err)
	}

	me, err := c.Me()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/workouts.csv", func(w http.ResponseWriter, req *http.Request) {
		workouts, err := c.WorkoutsCSV(me.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/csv")
		w.Write(workouts)
	})

	http.ListenAndServe(":9000", mux)
}
