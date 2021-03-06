package models

import (
	"fmt"
	"net/http"
	"time"
)

type Test struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Topic       string `json:"topic"`
	Score       uint64 `json:"score"`
	Passed      bool   `json:"is_passed,omitempty"`
}

type Stage struct {
	Name   string   `json:"name"`
	Status string   `json:"status"`
	Test   string   `json:"test,omitempty"`
	Output string   `json:"output,omitempty"`
	Run    *RunHash `json:"run,omitempty"`
	Cached bool     `json:"is_cached,omitempty"`
}

func (s *Stage) FillFromRun(stageName string, run *Run) {
	s.Name = fmt.Sprintf("%s::%s", stageName, run.Test)
	s.Status = run.Status
	s.Test = run.Test
	s.Output = run.Output
}

func (s *Stage) Success() bool {
	return s.Status == "success"
}

type Commit struct {
	Login  string   `json:"login"`
	Repo   string   `json:"repository"`
	Commit string   `json:"commit"`
	Status string   `json:"status"`
	Checks []*Stage `json:"checks"`
}

type Run struct {
	Hash     string `json:"hash"`
	Status   string `json:"status"`
	Output   string `json:"output"`
	Score    uint64 `json:"score"`
	Test     string `json:"test"`
	Baseline bool   `json:"baseline"`
}

type RunHash struct {
	Hash string `json:"hash"`
}

func (r *Run) CompareToBaseline(b *Run) {
	if r.Score == 0 {
		return
	}
	percent := r.Score * 1000 / b.Score
	humanPercent := float64(percent) / 10.
	r.Output = fmt.Sprintf("Performance: %.1f%% of baseline", humanPercent)
	if percent > 1200 {
		r.Status = "failure"
	} else {
		r.Status = "success"
	}
}

type Task struct {
	Id     string `json:"id"`
	Ref    string `json:"ref"`
	Url    string `json:"archive"`
	Stages []*Stage
	Runs   []*Run
}

func (t *Task) ReportSystemError(test string) {
	var name string
	if test == "" {
		name = "system"
	} else {
		name = fmt.Sprintf("test::%s", test)
	}
	t.Stages = append(t.Stages, &Stage{
		Name:   name,
		Status: "exception",
		Test:   test,
		Output: "System error. Reported to administrators.",
	})
}

type Stat struct {
	Login string `json:"login"`
	Score uint   `json:"score"`
	Count uint   `json:"count"`
}

type UserStats struct {
	Tests []*Test `json:"tests"`
	Score uint64  `json:"score"`
	Total uint64  `json:"total"`
}

type Course struct {
	Update time.Time `json:"updated_at,omitempty"`
	Ready  bool      `json:"is_ready"`
}

type AppInstallData struct {
	InstID uint64 `json:"installation_id"`
	State  string `json:"state"`
}

type AuthStage struct {
	Session string `json:"session,omitempty"`
	Url     string `json:"url,omitempty"`
}

func (as *AuthStage) SetAuthCookie(w http.ResponseWriter) {
	if as.Session == "" {
		return
	}
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{
		Name:     "session",
		Value:    as.Session,
		Expires:  expiration,
		HttpOnly: true,
		Path:     "/",
		// SameSite: http.SameSiteStrictMode,
		// Secure:   true,
	}
	http.SetCookie(w, &cookie)
}

type User struct {
	Id    uint64 `json:"id"`
	Login string `json:"login"`
	Repo  string `json:"repo"`
}
