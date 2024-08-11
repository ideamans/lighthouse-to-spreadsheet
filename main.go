package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/sheets/v4"
)

type LighthouseResult struct {
	Categories struct {
		Performance struct {
			Score     float64 `json:"score"`
			AuditRefs []struct {
				Id     string  `json:"id"`
				Weight float64 `json:"weight"`
			}
		} `json:"performance"`
	}
	Audits struct {
		LCP struct {
			Value float64 `json:"numericValue"`
			Score float64 `json:"score"`
		} `json:"largest-contentful-paint"`
		CLS struct {
			Value float64 `json:"numericValue"`
			Score float64 `json:"score"`
		} `json:"cumulative-layout-shift"`
		TBT struct {
			Value float64 `json:"numericValue"`
			Score float64 `json:"score"`
		} `json:"total-blocking-time"`
		FCP struct {
			Value float64 `json:"numericValue"`
			Score float64 `json:"score"`
		} `json:"first-contentful-paint"`
		FMP struct {
			Value float64 `json:"numericValue"`
			Score float64 `json:"score"`
		} `json:"first-meaningful-paint"`
		SI struct {
			Value float64 `json:"numericValue"`
			Score float64 `json:"score"`
		} `json:"speed-index"`
		TTFB struct {
			Value float64 `json:"numericValue"`
			Score float64 `json:"score"`
		} `json:"server-response-time"`
		TTI struct {
			Value float64 `json:"numericValue"`
			Score float64 `json:"score"`
		} `json:"interactive"`
	} `json:"audits"`
}

func ReadLighthouseResult(filePath string) (*LighthouseResult, error) {
	jsonFile, err := os.Open("./example/lighthouse.report.json")
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	jsonBuffer, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var result LighthouseResult
	if err := json.Unmarshal(jsonBuffer, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func loadGoogleCredential() (*jwt.Config, error) {
	homeDirPath, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	saFilePath := filepath.Join(homeDirPath, ".lighthouse-to-spreadsheet", "service-account.json")
	buffer, err := os.ReadFile(saFilePath)
	if err != nil {
		return nil, err
	}

	config, err := google.JWTConfigFromJSON(buffer, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func AppendSpreadSheet(spreadsheetId string, sheetName string, headerValues []interface{}, rowValues []interface{}) error {
	config, err := loadGoogleCredential()
	if err != nil {
		return err
	}

	ctx := context.Background()
	client := config.Client(ctx)

	srv, err := sheets.New(client)
	if err != nil {
		return err
	}

	rangeData := fmt.Sprintf("%s!A:A", sheetName)

	currentData, err := srv.Spreadsheets.Values.Get(spreadsheetId, rangeData).Do()
	if err != nil {
		return err
	}

	// データがまだない場合はヘッダを書き込み
	if len(currentData.Values) == 0 {
		var headerRange sheets.ValueRange
		headerRange.Values = append(headerRange.Values, headerValues)

		_, err = srv.Spreadsheets.Values.Append(spreadsheetId, rangeData, &headerRange).ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Do()
		if err != nil {
			return err
		}
	}

	var dataRange sheets.ValueRange
	dataRange.Values = append(dataRange.Values, rowValues)

	// データを追記
	_, err = srv.Spreadsheets.Values.Append(spreadsheetId, rangeData, &dataRange).ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		return err
	}

	return nil
}

type GitStatus struct {
	CommitId              string
	LastCommitMessage     string
	Branch                string
	Tags                  []string
	HasUncommittedChanges bool
}

func executeGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func GetGitStatus() (*GitStatus, error) {
	var status GitStatus

	// コミットIDを取得
	commitId, err := executeGitCommand("rev-parse", "HEAD")
	if err == nil {
		status.CommitId = commitId
	} else {
		fmt.Printf("failed to get commit id: %s\n", err)
	}

	// コミットメッセージ
	lastCommitMessage, err := executeGitCommand("log", "-1", "--pretty=%B")
	if err == nil {
		status.LastCommitMessage = lastCommitMessage
	} else {
		fmt.Printf("failed to get last commit message: %s\n", err)
	}

	// ブランチ
	branch, err := executeGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err == nil {
		status.Branch = branch
	} else {
		fmt.Printf("failed to get branch: %s\n", err)
	}

	// タグ
	tags, err := executeGitCommand("tag", "--points-at", "HEAD")
	if err == nil {
		status.Tags = strings.Split(tags, "\n")
	} else {
		fmt.Printf("failed to get tags: %s\n", err)
	}

	// コミットされていない変更の有無 (これはSTDERRに出力されるのでcombinedを使う)
	uncommittedChanges, err := executeGitCommand("status", "--porcelain")
	if err == nil {
		status.HasUncommittedChanges = uncommittedChanges != ""
	} else {
		fmt.Printf("failed to get uncommitted changes: %s\n", err)
	}

	return &status, nil
}

func GetCurrentDirBasename() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(dir)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	lighthouseResultPath := os.Getenv("LIGHTHOUSE_RESULT_PATH")
	spreadsheetId := os.Getenv("SPREADSHEET_ID")
	sheetName := os.Getenv("SHEET_NAME")

	flag.StringVar(&lighthouseResultPath, "lighthouse-result", lighthouseResultPath, "Path to lighthouse result")
	flag.StringVar(&spreadsheetId, "spreadsheet-id", spreadsheetId, "Spreadsheet ID")
	flag.StringVar(&sheetName, "sheet-name", sheetName, "Sheet name")
	flag.Parse()

	fmt.Printf("lighthouse-result: %s\n", lighthouseResultPath)
	fmt.Printf("spreadsheet-id: %s\n", spreadsheetId)
	fmt.Printf("sheet-name: %s\n", sheetName)
	return

	// Lighthouseの結果を読み込む
	lighthouseResult, err := ReadLighthouseResult(lighthouseResultPath)
	if err != nil {
		fmt.Println("Failed to read lighthouse result: ", err)
		os.Exit(1)
	}

	// Gitのステータスを取得
	gitStatus, err := GetGitStatus()
	if err != nil {
		// Gitのステータスは取得できなくても続行する
		fmt.Println("Failed to get git status: ", err)
	}

	// 現在のディレクトリ
	project := GetCurrentDirBasename()

	// 現在時刻
	now := time.Now()

	// スプレッドシートに書き込む
	headerValues := []interface{}{
		"プロジェクト", "時刻",
		"ブランチ", "コミットメッセージ", "タグ", "コミットID", "未コミットの変更",
		"パフォーマンススコア",
		"LCPスコア", "CLSスコア", "TBTスコア",
		"FCPスコア", "SIスコア",
		"LCP", "CLS", "TBT",
		"FCP", "FMP", "SI",
		"TTFB", "TTI",
	}

	tags := strings.Join(gitStatus.Tags, ", ")

	uncommitted := "N"
	if gitStatus.HasUncommittedChanges {
		uncommitted = "Y"
	}

	rowValues := []interface{}{
		project, now.Format("2006-01-02 15:04:05"),
		gitStatus.Branch, gitStatus.LastCommitMessage, tags, gitStatus.CommitId, uncommitted,
		lighthouseResult.Categories.Performance.Score * 100,
		lighthouseResult.Audits.LCP.Score * 100, lighthouseResult.Audits.CLS.Score * 100, lighthouseResult.Audits.TBT.Score * 100,
		lighthouseResult.Audits.FCP.Score * 100, lighthouseResult.Audits.SI.Score * 100,
		lighthouseResult.Audits.LCP.Value, lighthouseResult.Audits.CLS.Value, lighthouseResult.Audits.TBT.Value,
		lighthouseResult.Audits.FCP.Value, lighthouseResult.Audits.FMP.Value, lighthouseResult.Audits.SI.Value,
		lighthouseResult.Audits.TTFB.Value, lighthouseResult.Audits.TTI.Value,
	}

	if err := AppendSpreadSheet(spreadsheetId, sheetName, headerValues, rowValues); err != nil {
		fmt.Println("Failed to append spreadsheet: ", err)
		os.Exit(1)
	}
}
