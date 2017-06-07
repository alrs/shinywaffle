package main

import (
  "flag"
  "fmt"
  "log"
  "time"

  "github.com/hatstand/cc1101/config"
  "github.com/kidoman/embd"
  _ "github.com/kidoman/embd/host/rpi"
)

const (
  // Read/write flags.
  WRITE_SINGLE_BYTE = 0x00
  WRITE_BURST = 0x40
  READ_SINGLE_BYTE = 0x80
  READ_BURST = 0xc0

  BYTES_IN_RXFIFO = 0x7f
  RXFIFO = 0x3f
  OVERFLOW = 0x80

  CRC_OK = 0x80
  RSSI = 0
  LQI = 1
  RSSI_OFFSET = 74

  // Strobes
  SRES = 0x30  // Reset
  SRX = 0x34   // Set receive mode
  SIDLE = 0x36
  SFRX = 0x3a  // Flush RX FIFO buffer

  // Status Registers
  PARTNUM = 0xf0
  VERSION = 0xf1
  RXBYTES = 0x3b

  // Config Registers
  IOCFG2 = 0x00
  IOCFG1 = 0x01
  IOCFG0 = 0x02

  FIFOTHR = 0x03

  SYNC1 = 0x04
  SYNC0 = 0x05

  PKTLEN = 0x06
  PKTCTRL1 = 0x07
  PKTCTRL0 = 0x08

  ADDR = 0x09

  CHANNR = 0x0a
  FSCTRL1 = 0x0b
  FSCTRL0 = 0x0c

  FREQ2 = 0x0d
  FREQ1 = 0x0e
  FREQ0 = 0x0f

  MDMCFG4 = 0x10
  MDMCFG3 = 0x11
  MDMCFG2 = 0x12
  MDMCFG1 = 0x13
  MDMCFG0 = 0x14

  DEVIATN = 0x15

  MCSM2 = 0x16
  MCSM1 = 0x17
  MCSM0 = 0x18

  FOCCFG = 0x19
  BSCFG = 0x1a

  AGCCTRL2 = 0x1b
  AGCCTRL1 = 0x1c
  AGCCTRL0 = 0x1d

  WOREVT1 = 0x1e
  WOREVT0 = 0x1f
  WORCTRL = 0x20

  FREND1 = 0x21
  FREND0 = 0x22

  FSCAL3 = 0x23
  FSCAL2 = 0x24
  FSCAL1 = 0x25
  FSCAL0 = 0x26

  RCCTRL1 = 0x27
  RCCTRL0 = 0x28

  FSTEST = 0x29
  PTEST = 0x2a
  AGCTEST = 0x2b
  TEST2 = 0x2c
  TEST1 = 0x2d
  TEST0 = 0x2e
)

func convertRSSI(rssi int) int {
  if rssi >= 128 {
    return (rssi - 256) / 2 - RSSI_OFFSET
  } else {
    return rssi / 2 - RSSI_OFFSET
  }
}


type CC1101 struct {
  bus embd.SPIBus
}

func (cc1101 *CC1101) Strobe(address byte) error {
  data := []byte{address, 0x00}
  return cc1101.bus.TransferAndReceiveData(data)
}

func (cc1101 *CC1101) ReadSingleByte(address byte) (byte, error) {
  data := []byte{address | READ_SINGLE_BYTE, 0x00}
  err := cc1101.bus.TransferAndReceiveData(data)
  if err != nil {
    return 0x00, err
  }
  return data[1], nil
}

func (c *CC1101) ReadBurst(address byte, num byte) ([]byte, error) {
  var buf []byte

  for i:=byte(0); i < num + 1; i++ {
    addr := (address + i * 8) | READ_BURST
    buf = append(buf, addr)
  }

  err := c.bus.TransferAndReceiveData(buf)
  if err != nil {
    return nil, err
  }
  return buf[1:], nil
}

func (cc1101 *CC1101) WriteSingleByte(address byte, in byte) error {
  data := []byte{address | WRITE_SINGLE_BYTE, in}
  return cc1101.bus.TransferAndReceiveData(data)
}

func (cc1101 *CC1101) Reset() error {
  return cc1101.Strobe(SRES)
}

func (c *CC1101) Init() error {
  c.WriteSingleByte(FSCTRL1, config.FSCTRL1)
  c.WriteSingleByte(FSCTRL0, config.FSCTRL0)

  c.WriteSingleByte(FREQ2, config.FREQ2)
  c.WriteSingleByte(FREQ1, config.FREQ1)
  c.WriteSingleByte(FREQ0, config.FREQ0)

  c.WriteSingleByte(MDMCFG4, config.MDMCFG4)
  c.WriteSingleByte(MDMCFG3, config.MDMCFG3)
  c.WriteSingleByte(MDMCFG2, config.MDMCFG2)
  c.WriteSingleByte(MDMCFG1, config.MDMCFG1)
  c.WriteSingleByte(MDMCFG0, config.MDMCFG0)

  c.WriteSingleByte(CHANNR, config.CHANNR)
  c.WriteSingleByte(DEVIATN, config.DEVIATN)
  c.WriteSingleByte(FREND1, config.FREND1)
  c.WriteSingleByte(FREND0, config.FREND0)
  c.WriteSingleByte(MCSM0, config.MCSM0)
  c.WriteSingleByte(FOCCFG, config.FOCCFG)
  c.WriteSingleByte(BSCFG, config.BSCFG)

  c.WriteSingleByte(AGCCTRL2, config.AGCCTRL2)
  c.WriteSingleByte(AGCCTRL1, config.AGCCTRL1)
  c.WriteSingleByte(AGCCTRL0, config.AGCCTRL0)

  c.WriteSingleByte(FSCAL3, config.FSCAL3)
  c.WriteSingleByte(FSCAL2, config.FSCAL2)
  c.WriteSingleByte(FSCAL1, config.FSCAL1)
  c.WriteSingleByte(FSCAL0, config.FSCAL0)

  c.WriteSingleByte(FSTEST, config.FSTEST)
  c.WriteSingleByte(TEST2, config.TEST2)
  c.WriteSingleByte(TEST1, config.TEST1)
  c.WriteSingleByte(TEST0, config.TEST0)

  c.WriteSingleByte(IOCFG2, config.IOCFG2)
  c.WriteSingleByte(IOCFG0, config.IOCFG0)

  // Two status bytes appended to payload: RSSI LQI and CRC OK.
  c.WriteSingleByte(PKTCTRL1, config.PKTCTRL1)
  // No address check, data whitening off, CRC enable, variable length packets.
  c.WriteSingleByte(PKTCTRL0, config.PKTCTRL0)

  c.WriteSingleByte(ADDR, config.ADDR)
  // Max packet length 61 bytes.
  c.WriteSingleByte(PKTLEN, config.PKTLEN)
  return nil
}

func (cc1101 *CC1101) SelfTest() error {
  version, err := cc1101.ReadSingleByte(VERSION)
  if err != nil {
    return err
  }
  log.Printf("Version: 0x%x", version)
  partnum, err := cc1101.ReadSingleByte(PARTNUM)
  if err != nil {
    return err
  }
  log.Printf("Partnum: 0x%x", partnum)

  if version != 0x14 || partnum != 0x00 {
    return fmt.Errorf("Self test failed.\nGot Version: 0x%x Partnum: 0x%x", version, partnum)
  }
  return nil
}

func (cc1101 *CC1101) SetCarrierFrequency(freq int) error {
  if (freq == 868) {
    err := cc1101.WriteSingleByte(FREQ2, 0x21)
    if err != nil {
      return err
    }
    err = cc1101.WriteSingleByte(FREQ1, 0x62)
    if err != nil {
      return err
    }
    err = cc1101.WriteSingleByte(FREQ0, 0x76)
    if err != nil {
      return err
    }
    return nil
  } else {
    return fmt.Errorf("Frequency %dMHz not supported.", freq)
  }
}

func (cc1101 *CC1101) SetRx() error {
  return cc1101.Strobe(SRX)
}

func (c *CC1101) SetIdle() error {
  return c.Strobe(SIDLE)
}

func (c *CC1101) FlushRx() {
  c.Strobe(SIDLE)
  time.Sleep(50 * time.Microsecond)
  c.Strobe(SFRX)
}

func (c *CC1101) Receive() ([]byte, error) {
  rxbytes, err := c.ReadSingleByte(RXBYTES)
  if err != nil {
    return nil, err
  }
  log.Printf("RXBYTES: 0x%x\n", rxbytes)
  // Flush RX buffer.
  defer c.FlushRx()

  if rxbytes & OVERFLOW > 0 {
    return nil, fmt.Errorf("FIFO Overflow")
  }

  if rxbytes & BYTES_IN_RXFIFO > 0 {
    log.Printf("Bytes in buffer: %d", rxbytes & BYTES_IN_RXFIFO)
    numBytes, err := c.ReadSingleByte(RXFIFO)
    if err != nil {
      return nil, err
    }
    log.Printf("Receiving %d bytes", numBytes)
    var recv []byte
    if numBytes > 0 {
      recv, err = c.ReadBurst(RXFIFO, numBytes)
      if err != nil {
        return nil, err
      }
    }
    status, err := c.ReadBurst(RXFIFO, 2)
    if err != nil {
      return nil, err
    }

    log.Printf("Status RSSI:   %ddBm\n", convertRSSI(int(status[RSSI])))
    log.Printf("Status LQI:    %d\n", status[LQI] & 0x7f)
    log.Printf("Status CRC OK: %d\n", (status[LQI] & CRC_OK) >> 7)
    return recv, nil
  } else {
    return []byte{}, nil
  }
}

func main() {
  flag.Parse()
  err := embd.InitSPI()
  if err != nil {
    panic(err)
  }
  defer embd.CloseSPI()

  err = embd.InitGPIO()
  if err != nil {
    panic(err)
  }
  defer embd.CloseGPIO()

  gdo0, err := embd.NewDigitalPin(24)
  if err != nil {
    panic(err)
  }
  defer gdo0.Close()

  gdo2, err := embd.NewDigitalPin(25)
  if err != nil {
    panic(err)
  }
  defer gdo2.Close()

  gdo0.SetDirection(embd.In)
  gdo2.SetDirection(embd.In)

  bus := embd.NewSPIBus(embd.SPIMode0, 0, 50000, 8, 0)
  defer bus.Close()

  cc1101 := CC1101{
    bus: bus,
  }
  cc1101.Reset()
  cc1101.SelfTest()
  cc1101.Init()

  for {
    cc1101.SetIdle()
    cc1101.SetRx()
    log.Print("Waiting for packets...")
    for {
      flag, err := gdo0.Read()
      if err != nil {
        panic(err)
      }
      if flag > 0 {
        break
      }
    }
    log.Print("Received something!")
    time.Sleep(time.Microsecond * 50)
    recv, err := cc1101.Receive()
    if err != nil {
      log.Println("Failed to receive: ", err)
    } else {
      log.Printf("Received: %v\n", recv)
    }
  }
}

