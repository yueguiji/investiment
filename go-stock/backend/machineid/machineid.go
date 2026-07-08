package machineid

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var buildKey string

func Init(key string) {
	buildKey = key
}

func GetMachineId() string {
	salt := buildKey
	if salt == "" {
		salt = "cc1e0d684e32f176c56ff1fcf384dcd9"
	}

	id := systemMachineID()
	if id == "" {
		id = fallbackMachineID()
	}
	sum := sha256.Sum256([]byte(salt + ":" + id))
	return hex.EncodeToString(sum[:])
}

func systemMachineID() string {
	if runtime.GOOS != "windows" {
		return ""
	}
	out, err := exec.Command("reg", "query", `HKLM\SOFTWARE\Microsoft\Cryptography`, "/v", "MachineGuid").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "MachineGuid") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			return fields[len(fields)-1]
		}
	}
	return ""
}

func fallbackMachineID() string {
	hostname, _ := os.Hostname()
	if strings.TrimSpace(hostname) == "" {
		hostname = runtime.GOOS
	}
	return hostname
}
