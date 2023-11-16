package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
	"unsafe"
	// "net/ipv4"
)

const (
	// tun设备文件
	tunDeviceFile = "/dev/net/tun"
	// tun设备名称
	tunDeviceName = "tun1"
	// 默认MTU为1500，此处用作数据buffer大小
	// 一个IP头（20字节）和一个UDP头（8）字节。如果f设置1500，
	// 加上IP和UDP头的28字节数据，到达宿主机eth0的时候最大报文会超过eth0的MTU，eth0会把该数据包丢弃
	defaultMTU = 2000
	udpPort    = 1234
	tcpPort    = 1234
)

// var dstUDPHost = flag.String("d", "127.0.0.1:8285", "destination UDP host")

type ifReq struct {
	name  [16]byte
	flags uint16
}

func main() {
	flag.Parse()
	// 初始化tun设备

	// conn, sErr := UDPServer()
	// if sErr != nil {
	// 	log.Printf("initSocket error: %s", sErr.Error())
	// 	return
	// }
	// defer conn.Close()
	// log.Printf("initSocket listen 127.0.0.1 with socket success")

	tun, tErr := InitTunDevice()
	if tErr != nil {
		log.Printf("initTunDevice error: %s", tErr.Error())
		return
	}
	// defer tun.Close()
	log.Printf("initTunDevice tun with file %s success", tunDeviceFile)

	tcpServerHost := fmt.Sprintf(":%d", tcpPort)

	// conn, err := net.ListenTCP("tcp4", &net.TCPAddr{
	// 	IP:   net.IPv4(127, 0, 0, 1),
	// 	Port: tcpPort,
	// })
	conn, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatalf("net.Listen error: %s", err.Error())
	}
	defer conn.Close()
	log.Printf("tcp listen on: %s", tcpServerHost)
	// log.Printf("udp lis: %s", updServerHost)
	for {
		tcpconn, err := conn.Accept()
		if err != nil {
			log.Printf("tcp accetp err: %s", err.Error())
		}
		defer tcpconn.Close()
		go Socket2Tun(tcpconn, tun)
		go Tun2Socket(tun, tcpconn)

	}

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

func Socket2Tun(conn net.Conn, tun *os.File) {

	for {
		buffer := make([]byte, defaultMTU)
		blen, err := conn.Read(buffer)

		if err != nil {
			fmt.Printf("conn.Read err:[%v]\n", err)
		}
		// defer conn.Close()
		log.Printf("---ReadNet---read %d bytes data from tcp", blen)

		if blen > 0 {
			tlen, err := tun.Write(buffer[:blen])
			if err != nil {
				log.Printf("handler write data to tun device error: %s", err.Error())
				// continue
			}
			log.Printf("---Write2Tun--- handler write %d bytes data to tun device", tlen)
		}

	}
}

func Tun2Socket(tun *os.File, conn net.Conn) {
	for {
		buffer := make([]byte, defaultMTU)

		tn, tErr := tun.Read(buffer)
		if tErr != nil {
			log.Printf("---ReadTun---read data from tun error: %s", tErr.Error())
			return
		}
		// log.Panicf("read %d byte data from tun device", tn)
		log.Printf("read %d byte data from tun device", tn)

		//	data := buffer[:tn]
		// conn.remoteAddr()
		if tn > 0 {
			cn, cErr := conn.Write(buffer[:tn])
			// return conn.Write(data)
			if cErr != nil {
				log.Printf("write data to tcp device error: %s", cErr.Error())
				return
			}
			log.Printf("---Write2Net---write %d bytes data to tcp device", cn)

		}

	}

}
