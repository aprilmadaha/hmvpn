package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
	"unsafe"
)

const (
	// tun设备文件
	tunDeviceFile = "/dev/net/tun"
	// tun设备名称
	tunDeviceName = "tun1"
	// 默认MTU为1500，此处用作数据buffer大小
	// 一个IP头（20字节）和一个UDP头（8）字节。如果f设置1500，
	// 加上IP和UDP头的28字节数据，到达宿主机eth0的时候最大报文会超过eth0的MTU，eth0会把该数据包丢弃
	defaultMTU = 1472
	udpPort    = 8285
)

var dstUDPHost = flag.String("d", "127.0.0.1:8285", "destination UDP host")

type ifReq struct {
	name  [16]byte
	flags uint16
}

func main() {
	flag.Parse()
	// 初始化tun设备

	conn, sErr := SendUDPData()
	if sErr != nil {
		log.Printf("initSocket error: %s", sErr.Error())
		return
	}
	defer conn.Close()
	log.Printf("send data success")

	tun, tErr := InitTunDevice()
	if tErr != nil {
		log.Printf("initTunDevice error: %s", tErr.Error())
		return
	}
	defer tun.Close()
	log.Printf("initTunDevice tun with file %s success", tunDeviceFile)

	Tun2Socket(tun, conn)
}

func InitTunDevice() (*os.File, error) {
	// 打开tun设备文件
	tun, err := os.OpenFile(tunDeviceFile, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("os.OpenFile error: %s", err.Error())
	}

	// ioctl设置
	var ir = ifReq{
		flags: syscall.IFF_TUN | syscall.IFF_NO_PI,
	}
	copy(ir.name[:], tunDeviceName)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, tun.Fd(), syscall.TUNSETIFF, uintptr(unsafe.Pointer(&ir)))
	if errno != 0 {
		return nil, fmt.Errorf("ioctl error: expect 0 but got %d", errno)
	}
	log.Printf("ioctl success")

	return tun, nil
}

// func Socket2Tun(conn *net.UDPConn, tun *os.File) {
// 	var buffer = make([]byte, defaultMTU)
// 	for {
// 		cn, remoteAddr, err := conn.ReadFromUDP(buffer)
// 		if err != nil {
// 			fmt.Printf("conn.ReadFromUDP err:[%v]\n", err)
// 		}
// 		// defer conn.Close()
// 		log.Printf("read %d bytes data from udp %s", cn, remoteAddr)

// 		// 接收数据
// 		// go handler(buffer[:n])

// 		data := buffer[:cn]
// 		// _, err = conn.WriteToUDP(data, remoteAddr)
// 		tn, err := tun.Write(data)
// 		if err != nil {
// 			log.Printf("handler write data to tun device error: %s", err.Error())
// 			return
// 		}
// 		log.Printf("handler write %d bytes data to tun device", tn)
// 	}
// }

func Tun2Socket(tun *os.File, conn *net.UDPConn) {
	buffer := make([]byte, defaultMTU)
	for {
		tn, tErr := tun.Read(buffer)
		if tErr != nil {
			log.Printf("read data from tun error: %s", tErr.Error())
			return
		}
		log.Panicf("read %d byte data from tun device", tn)

		// data := buffer[:tn]
		// conn.remoteAddr()
		cn, cErr := conn.Write(buffer)
		// return conn.Write(data)
		if cErr != nil {
			log.Printf("handler write data to tun device error: %s", cErr.Error())
			return
		}
		log.Printf("handler write %d bytes data to tun device", cn)

	}
}

// UDPServer 接收UDP数据
// func UDPServer(handler func(data []byte))
// func UDPServer() (*net.UDPConn, error) {
// 	updServerHost := fmt.Sprintf(":%d", udpPort)
// 	conn, err := net.ListenUDP("udp", &net.UDPAddr{
// 		IP:   net.IPv4(0, 0, 0, 0),
// 		Port: udpPort,
// 	})
// 	if err != nil {
// 		log.Fatalf("net.Listen error: %s", err.Error())
// 	}
// 	// defer conn.Close()

// 	log.Printf("udp listen on: %s", updServerHost)

// 	return conn, nil
// }

func SendUDPData() (*net.UDPConn, error) {
	serverAddr, err := net.ResolveUDPAddr("udp", *dstUDPHost)
	if err != nil {
		log.Fatalln("failed to resolve server addr:", err)
	}
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Errorf("net.Dial error: %s", err.Error())
	}
	log.Printf("dialUDP to %s success serverAddr", serverAddr)
	log.Printf("dialUDP to %s success conn.RemoteAddr()", conn.RemoteAddr())
	// return conn.Write(data)
	return conn, nil
}
