package garden

import (
	uuid "github.com/satori/go.uuid"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

// NewUuid 生成UUID
func NewUuid() string {
	return uuid.NewV4().String()
}

//ParseUuid 解析UUID
func ParseUuid(s string) bool {
	_, err := uuid.FromString(s)
	if err != nil {
		return false
	}
	return true
}

// ToDatetimeMillion time转换为日期时间格式(毫秒)
func ToDatetimeMillion(t time.Time) string {
	s := t.Format("2006-01-02 15:04:05.000")
	return s
}

// ToDatetime time转换为日期时间格式(秒)
func ToDatetime(t time.Time) string {
	s := t.Format("2006-01-02 15:04:05")
	return s
}

// Timing 计算两time时间间隔
func Timing(t1 time.Time, t2 time.Time) string {
	return t2.Sub(t1).String()
}

// IsNum 判断字符串是否数字
func IsNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// GetIntranetIp 获取本机内网IP
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

// GetOutboundIP 获取本机内网IP：当机器上存在多个IP接口时，这里有一个更好的解决方案来检索首选的出站IP地址
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

// CreateDir 创建文件夹
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

// PathExists 判断路径是否存在
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
