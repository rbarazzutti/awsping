package awsping

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// AWSRegion description of the AWS EC2 region
type AWSRegion struct {
	Name      string
	Code      string
	DomainExt string
	Service   string
	Latencies []Measure
	Error     error
}

// CheckLatencyICMP Test Latency via ICMP
func (r *AWSRegion) CheckLatencyICMP(seq int) {
	const DataSize = 56

	targetIP := r.GetTarget().IPAddr
	shortPid := os.Getpid() & 0xffff
	seq = seq & 0xffff

	buf := make([]byte, DataSize)
	for i := 0; i < DataSize; i++ {
		buf[i] = byte(i)
	}

	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		println("Failed to set up an ICMP endpoint.")
		println("note: ICMP ping requires root access or cap_net_raw=ep capability")
		os.Exit(1)
	}

	defer c.Close()

	{
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID:   shortPid,
				Seq:  seq,
				Data: buf,
			},
		}

		startTimeStampMicro := time.Now().UnixMicro()

		binary.BigEndian.PutUint32(buf, uint32(startTimeStampMicro/1e6))
		binary.BigEndian.PutUint32(buf[4:], uint32(startTimeStampMicro%1e6))

		wb, _ := wm.Marshal(nil)

		c.WriteTo(wb, targetIP)
	}

	rb := make([]byte, 1500)
	var delay = FailedMeasure
	c.SetReadDeadline(time.Now().Add(time.Second))
	for {
		n, peer, err := c.ReadFrom(rb)
		if err != nil {
			break
		}
		if rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), rb[:n]); err == nil {

			if rm.Type == ipv4.ICMPTypeEchoReply && peer.String() == targetIP.String() {
				body, _ := rm.Body.(*icmp.Echo)
				if body.ID == shortPid && body.Seq == seq {
					delay = Measure(time.Since(time.Unix(int64(binary.BigEndian.Uint32(body.Data)), int64(binary.BigEndian.Uint32(body.Data[4:]))*1e3)))
					break
				}
			}
		}
	}

	r.Latencies = append(r.Latencies, delay)

	r.Error = err
}

// AWSTarget describes aws region network details (host, ip)
type AWSTarget struct {
	Hostname string
	IPAddr   *net.IPAddr
}

// CheckLatencyHTTP Test Latency via HTTP

func (r *AWSRegion) CheckLatencyHTTP(https bool) {
	proto := "http"
	if https {
		proto = "https"
	}
	url := fmt.Sprintf("%s://%s/ping?x=%s", proto, r.GetTarget().Hostname, mkRandoString(13))

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		r.Error = err
	}
	req.Header.Set("User-Agent", useragent)
	start := time.Now()
	resp, err := client.Do(req)
	r.Latencies = append(r.Latencies, Measure(time.Since(start)))
	defer resp.Body.Close()

	r.Error = err
}

func (r *AWSRegion) GetTarget() AWSTarget {

	hostname := fmt.Sprintf("%s.%s.amazonaws.com%s", r.Service, r.Code, r.DomainExt)
	ipAddr, err := net.ResolveIPAddr("ip4", hostname)
	if err != nil {
		r.Error = err
	}

	return AWSTarget{
		Hostname: hostname,
		IPAddr:   ipAddr,
	}

}

// CheckLatencyTCP Test Latency via TCP
func (r *AWSRegion) CheckLatencyTCP() {

	tcpAddr :=
		net.TCPAddr{
			IP:   r.GetTarget().IPAddr.IP,
			Port: 80}

	start := time.Now()
	conn, err := net.DialTCP("tcp", nil, &tcpAddr)
	if err != nil {
		r.Error = err
		return
	}
	r.Latencies = append(r.Latencies, Measure(time.Since(start)))
	defer conn.Close()

	r.Error = err
}

// GetLatency returns Latency in ms
func (r *AWSRegion) GetAvgLatency() Measure {
	sum := Measure(0)
	count := int64(0)
	for _, l := range r.Latencies {
		if l.isValid() {
			count++
			sum += l
		}
	}
	if count > 0 {
		return Measure(int64(sum) / count)
	} else {
		return FailedMeasure
	}
}

// AWSRegions slice of the AWSRegion
type AWSRegions []AWSRegion

func (rs AWSRegions) Len() int {
	return len(rs)
}

func (rs AWSRegions) Less(i, j int) bool {
	a := rs[i].GetAvgLatency()
	b := rs[j].GetAvgLatency()

	if a.isValid() && b.isValid() && a != b {
		return a < b
	} else if a.isValid() && !b.isValid() {
		return true
	} else if !a.isValid() && b.isValid() {
		return false
	} else {
		return rs[i].Code < rs[j].Code
	}
}

func (rs AWSRegions) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

// GetRegions returns a list of regions
func GetRegions(service string) AWSRegions {
	return AWSRegions{
		{Service: service, Name: "US-East (N. Virginia)", Code: "us-east-1"},
		{Service: service, Name: "US-East (Ohio)", Code: "us-east-2"},
		{Service: service, Name: "US-West (N. California)", Code: "us-west-1"},
		{Service: service, Name: "US-West (Oregon)", Code: "us-west-2"},
		{Service: service, Name: "Canada (Central)", Code: "ca-central-1"},
		{Service: service, Name: "Europe (Ireland)", Code: "eu-west-1"},
		{Service: service, Name: "Europe (Frankfurt)", Code: "eu-central-1"},
		{Service: service, Name: "Europe (London)", Code: "eu-west-2"},
		{Service: service, Name: "Europe (Milan)", Code: "eu-south-1"},
		{Service: service, Name: "Europe (Paris)", Code: "eu-west-3"},
		{Service: service, Name: "Europe (Stockholm)", Code: "eu-north-1"},
		{Service: service, Name: "Africa (Cape Town)", Code: "af-south-1"},
		{Service: service, Name: "Asia Pacific (Osaka)", Code: "ap-northeast-3"},
		{Service: service, Name: "Asia Pacific (Hong Kong)", Code: "ap-east-1"},
		{Service: service, Name: "Asia Pacific (Tokyo)", Code: "ap-northeast-1"},
		{Service: service, Name: "Asia Pacific (Seoul)", Code: "ap-northeast-2"},
		{Service: service, Name: "Asia Pacific (Singapore)", Code: "ap-southeast-1"},
		{Service: service, Name: "Asia Pacific (Mumbai)", Code: "ap-south-1"},
		{Service: service, Name: "Asia Pacific (Sydney)", Code: "ap-southeast-2"},
		{Service: service, Name: "South America (São Paulo)", Code: "sa-east-1"},
		{Service: service, Name: "Middle East (Bahrain)", Code: "me-south-1"},
		{Service: service, Name: "AWS GovCloud (US-East)", Code: "us-gov-east-1"},
		{Service: service, Name: "AWS GovCloud (US-West)", Code: "us-gov-west-1"},
		{Service: service, Name: "China (Beijing)", Code: "cn-north-1", DomainExt: ".cn"},
		{Service: service, Name: "China (Ningxia)", Code: "cn-northwest-1", DomainExt: ".cn"},
	}

}

// CalcLatency returns list of aws regions sorted by Latency
func CalcLatency(repeats int, useHTTP bool, useHTTPS bool, useICMP bool, service string) *AWSRegions {
	regions := GetRegions(service)

	for n := 1; n <= repeats; n++ {
		var wg sync.WaitGroup
		wg.Add(len(regions))
		for i := range regions {
			go func(i int) {
				defer wg.Done()
				if useHTTP || useHTTPS {
					regions[i].CheckLatencyHTTP(useHTTPS)
				} else if useICMP {
					regions[i].CheckLatencyICMP(n)
				} else {
					regions[i].CheckLatencyTCP()
				}
			}(i)
		}
		wg.Wait()
	}

	sort.Sort(regions)
	return &regions
}
