package resolver

import (
	"net"
	"os"
	"runtime"
	"strings"
)

const (
	UNKNOWN = 0
	EMPTY   = 10
	COMMENT = 20
	ADDRESS = 30
)

func LoadHosts() (hosts map[string]net.IP, err error) {
	hostFile := "/etc/hosts"

	if runtime.GOOS == "windows" {
		hostFile = `C:\Windows\System32\Drivers\etc\hosts`
	}
	hfl, err := ParseHosts(hostFile)
	if err != nil {
		return
	}
	hosts = make(map[string]net.IP, len(hosts))
	for _, hs := range hfl {
		for _, h := range hs.Hostnames {
			hosts[h] = net.ParseIP(hs.Address)
		}
	}

	return
}

type HostFileLine struct {
	OriginalLineNum int
	LineType        int
	Address         string
	Parts           []string
	Hostnames       []string
	Raw             string
	Trimmed         string
	Comment         string
}

func ParseHosts(path string) ([]HostFileLine, error) {
	input, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseHostsFromString(string(input))
}

func ParseHostsFromString(input string) ([]HostFileLine, error) {
	inputNormalized := strings.Replace(input, "\r\n", "\n", -1)

	dataLines := strings.Split(inputNormalized, "\n")
	//remove extra blank line at end that does not exist in /etc/hosts file
	dataLines = dataLines[:len(dataLines)-1]

	hostFileLines := make([]HostFileLine, len(dataLines))

	// trim leading and trailing whitespace
	for i, l := range dataLines {
		curLine := &hostFileLines[i]
		curLine.OriginalLineNum = i
		curLine.Raw = l

		// trim line
		curLine.Trimmed = strings.TrimSpace(l)

		// check for comment
		if strings.HasPrefix(curLine.Trimmed, "#") {
			curLine.LineType = COMMENT
			continue
		}

		if curLine.Trimmed == "" {
			curLine.LineType = EMPTY
			continue
		}

		curLineSplit := strings.SplitN(curLine.Trimmed, "#", 2)
		if len(curLineSplit) > 1 {
			curLine.Comment = curLineSplit[1]
		}
		curLine.Trimmed = curLineSplit[0]

		curLine.Parts = strings.Fields(curLine.Trimmed)

		if len(curLine.Parts) > 1 {
			curLine.LineType = ADDRESS
			curLine.Address = strings.ToLower(curLine.Parts[0])
			// lower case all
			for _, p := range curLine.Parts[1:] {
				curLine.Hostnames = append(curLine.Hostnames, strings.ToLower(p))
			}

			continue
		}

		// if we can't figure out what this line is
		// at this point mark it as unknown
		curLine.LineType = UNKNOWN

	}

	return hostFileLines, nil
}
