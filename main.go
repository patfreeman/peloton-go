package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
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

type WorkoutRecord struct {
	WorkoutTimestamp	string	`json:"workout_timestamp"`
	Live			string	`json:"live"`
	InstructorName		string	`json:"instructor_name"`
	Length			int	`json:"length"`
	FitnessDiscipline	string	`json:"fitness_discipline"`
	Type			string	`json:"type"`
	Title			string	`json:"title"`
	ClassTimestamp		string	`json:"class_timestamp"`
	TotalOutput		int	`json:"total_output"`
	AvgWatts		int	`json:"avg_watts"`
	AvgResistance		string	`json:"avg_resistance"`
	AvgCadence		int	`json:"avg_cadence"`
	AvgSpeed		string	`json:"avg_speed"`
	Distance		string	`json:"distance"`
	CaloriesBurned		string	`json:"calories_burned"`
	AvgHeartrate		string	`json:"avg_heartrate"`
	AvgIncline		string	`json:"avg_incline"`
	AvgPace			string	`json:"avg_pace"`
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

func createWorkoutJson(data []string) []WorkoutRecord {
	var workoutJson []WorkoutRecord
	for i, line := range data {
		if i > 0 { // omit header
			var rec WorkoutRecord
			workoutLine := strings.Split(line, ",")
			if workoutLine[0] == "" { // omit blank records
				continue
			}
			for j, field := range workoutLine {
				if j == 0 {
					rec.WorkoutTimestamp = field
				} else if j == 1 {
					rec.Live = field
				} else if j == 2 {
					rec.InstructorName = field
				} else if j == 3 {
					var err error
					rec.Length, err = strconv.Atoi(field)
					if err != nil { continue }
				} else if j == 4 {
					rec.FitnessDiscipline = field
				} else if j == 5 {
					rec.Type = field
				} else if j == 6 {
					rec.Title = field
				} else if j == 7 {
					rec.ClassTimestamp = field
				} else if j == 8 {
					var err error
					rec.TotalOutput, err = strconv.Atoi(field)
					if err != nil { continue }
				} else if j == 9 {
					var err error
					rec.AvgWatts, err = strconv.Atoi(field)
					if err != nil { continue }
				} else if j == 10 {
					rec.AvgResistance = field
				} else if j == 11 {
					var err error
					rec.AvgCadence, err = strconv.Atoi(field)
					if err != nil { continue }
				} else if j == 12 {
					rec.AvgSpeed = field
				} else if j == 13 {
					rec.Distance = field
				} else if j == 14 {
					rec.CaloriesBurned = field
				} else if j == 15 {
					rec.AvgHeartrate = field
				} else if j == 16 {
					rec.AvgIncline = field
				} else if j == 17 {
					rec.AvgPace = field
				}
			}
			workoutJson = append(workoutJson, rec)
		}
	}
	return workoutJson
}

func (c *Client) WorkoutsCSV(userID string) ([]byte, error) {
	res, err := c.client.Get(fmt.Sprintf("%s/api/user/%s/workout_history_csv", baseURL, userID))

	if err != nil {
		return []byte{}, err
	}

	return ioutil.ReadAll(res.Body)
}

func handleRequest(ctx context.Context, event events.SQSEvent) (string, error) {
	c := NewClient()
	err := c.Login(os.Getenv("USERNAME"), os.Getenv("PASSWORD"))
	if err != nil {
		log.Fatal(err)
	}

	me, err := c.Me()
	if err != nil {
		log.Fatal(err)
	}

	workouts, err := c.WorkoutsCSV(me.ID)
	if err != nil {
		log.Fatal(err)
		return err.Error(), nil
	}

	workoutStrings := strings.Split(string(workouts),"\n")
	workoutJson := createWorkoutJson(workoutStrings)
	jsonData, err := json.MarshalIndent(workoutJson, "", "  ")
	if err != nil {
		log.Fatal(err)
		return err.Error(), nil
	}

	return string(jsonData), nil
}

func main() {
  runtime.Start(handleRequest)
}
