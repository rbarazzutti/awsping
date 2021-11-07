package awsping

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var (
	Version   = "1.0.0"
	github    = "https://github.com/ekalinin/awsping"
	useragent = fmt.Sprintf("AwsPing/%s (+%s)", Version, github)
)

const (
	ShowOnlyRegions = -1
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func mkRandoString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// LatencyOutput prints data into console
type LatencyOutput struct {
	Level   int
	Repeats int
}

func (lo *LatencyOutput) show(regions *AWSRegions) {
	for _, r := range *regions {
		fmt.Printf("%-15s %-20s\n", r.Code, r.Name)
	}
}

func (lo *LatencyOutput) show0(regions *AWSRegions) {
	for _, r := range *regions {
		fmt.Printf("%-25s %20s\n", r.Name, r.GetLatency().toStr())
	}
}

func (lo *LatencyOutput) show1(regions *AWSRegions) {
	outFmt := "%5v %-15s %-30s %20s\n"
	fmt.Printf(outFmt, "", "Code", "Region", "Latency")
	for i, r := range *regions {
		fmt.Printf(outFmt, i, r.Code, r.Name, r.GetLatency().toStr())
	}
}

func (lo *LatencyOutput) show2(regions *AWSRegions) {
	// format
	outFmt := "%5v %-15s %-25s"
	outFmt += strings.Repeat(" %15s", lo.Repeats) + " %15s\n"
	// header
	outStr := []interface{}{"", "Code", "Region"}
	for i := 0; i < lo.Repeats; i++ {
		outStr = append(outStr, "Try #"+strconv.Itoa(i+1))
	}
	outStr = append(outStr, "Avg Latency")

	// show header
	fmt.Printf(outFmt, outStr...)

	// each region stats
	for i, r := range *regions {
		outData := []interface{}{strconv.Itoa(i), r.Code, r.Name}
		for n := 0; n < lo.Repeats; n++ {
			outData = append(outData, r.GetLatency().toStr())
		}
		outData = append(outData, r.GetLatency().toStr())
		fmt.Printf(outFmt, outData...)
	}
}

// Show print data
func (lo *LatencyOutput) Show(regions *AWSRegions) {
	switch lo.Level {
	case ShowOnlyRegions:
		lo.show(regions)
	case 0:
		lo.show0(regions)
	case 1:
		lo.show1(regions)
	case 2:
		lo.show2(regions)
	}
}

type Measure time.Duration

const FailedMeasure = Measure(-1)

func (m Measure) isValid() bool {
	return int64(time.Duration(m)) >= int64(0)
}

func (m Measure) toStr() string {
	if m.isValid() {
		return fmt.Sprintf("%.1f ms", float64(time.Duration(m).Microseconds())/1000)
	} else {
		return "timeout"
	}
}
