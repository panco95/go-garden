package utils

import (
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"
)

func NewUuid() string {
	return uuid.NewV4().String()
}

func ParseUuid(s string) bool {
	_, err := uuid.FromString(s)
	if err != nil {
		return false
	}
	return true
}

func ToDatetimeMillion(t time.Time) string {
	s := t.Format("2006-01-02 15:04:05.000")
	return s
}

func ToDatetime(t time.Time) string {
	s := t.Format("2006-01-02 15:04:05")
	return s
}

func Timing(t1 time.Time, t2 time.Time) string {
	return t2.Sub(t1).String()
}

func IsNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func GetIntranetIp() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}

		}
	}
	return "127.0.0.1"
}

func GetOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

func CreateDir(path string) error {
	exists, err := PathExists(path)
	if err != nil {
		return err
	}
	if !exists {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ReadFile(filepath string) ([]byte, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func WriteFile(filepath string, data []byte) error {
	if err := ioutil.WriteFile(filepath, data, 0777); err != nil {
		return nil
	}
	return nil
}
