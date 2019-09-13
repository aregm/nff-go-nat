// Copyright 2017 Intel Corporation.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"

	"github.com/intel-go/nff-go/flow"

	"github.com/intel-go/nff-go-nat/nat"
)

type DumpControlArray [nat.DirKNI + 1]bool

func (drc *DumpControlArray) String() string {
	return ""
}

func (drc *DumpControlArray) Set(value string) error {
	fmt.Println("Input: ", value)
	for _, c := range value {
		if c == 'd' {
			(*drc)[nat.DirDROP] = true
		} else if c == 't' {
			(*drc)[nat.DirSEND] = true
		} else if c == 'k' {
			(*drc)[nat.DirKNI] = true
		} else {
			return fmt.Errorf("Bad dump control flag character: \"%v\"", c)
		}
	}
	return nil
}

func main() {
	var dumpControl DumpControlArray
	// Parse arguments
	cores := flag.String("cores", "", "Specify CPU cores to use.")
	configFile := flag.String("config", "config.json", "Specify config file name.")
	flag.BoolVar(&nat.NoCalculateChecksum, "nocsum", false, "Specify whether to calculate checksums in modified packets.")
	flag.BoolVar(&nat.NoHWTXChecksum, "nohwcsum", false, "Specify whether to use hardware offloading for checksums calculation (requires -csum).")
	noscheduler := flag.Bool("no-scheduler", false, "Disable scheduler.")
	setKniIP := flag.Bool("set-kni-IP", false, "Set IP addresses specified in config file to created KNI interfaces. Do not use if your system uses Network Manager! Use Network Manager configurations instead.")
	bringUpKniInterfaces := flag.Bool("bring-up-kni", false, "Set IP addresses specified in config file to created KNI interfaces. Do not use if your system uses Network Manager! Use Network Manager configurations instead.")
	dpdkLogLevel := flag.String("dpdk", "--log-level=0", "Passes an arbitrary argument to dpdk EAL.")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Var(&dumpControl, "dump", `Enable dump pcap output in a form of letter flags,
e.g. "-dump d" or "-dump dtk":
    d means to trace dropped packets,
    t means to trace translated (normally sent) packets,
    k means to trace packets that were sent to KNI interface.`)
	schedulerInterval := flag.Uint("scheduler-interval", 500, "Set scheduler interval in ms. Lower values allow faster reaction to changing traffic but increase scheduling overhead.")
	sendCPUCoresPerPort := flag.Int("send-threads", 1, "Number of CPU cores to be occupied by Send routines.")
	tXQueuesNumberPerPort := flag.Int("tx-queues", 4, "Number of transmit queues to use on network card.")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		flow.CheckFatal(err)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	nat.DumpEnabled = dumpControl

	// Set up reaction to SIGINT (Ctrl-C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Read config
	flow.CheckFatal(nat.ReadConfig(*configFile, *setKniIP, *bringUpKniInterfaces))

	// Init NFF-GO system at 16 available cores
	nffgoconfig := flow.Config{
		CPUList:               *cores,
		HWTXChecksum:          !nat.NoHWTXChecksum,
		DPDKArgs:              []string{*dpdkLogLevel},
		DisableScheduler:      *noscheduler,
		NeedKNI:               nat.NeedKNI,
		SchedulerInterval:     *schedulerInterval,
		SendCPUCoresPerPort:   *sendCPUCoresPerPort,
		TXQueuesNumberPerPort: *tXQueuesNumberPerPort,
	}

	flow.CheckFatal(flow.SystemInit(&nffgoconfig))

	offloadingAvailable := nat.CheckHWOffloading()
	if !nat.NoHWTXChecksum && !offloadingAvailable {
		println("Warning! Requested hardware offloading is not available on all ports. Falling back to software checksum calculation.")
		nat.NoHWTXChecksum = true
		flow.SetUseHWCapability(flow.HWTXChecksumCapability, false)
	}

	// Initialize flows and necessary state
	nat.InitFlows()

	// Start GRPC server
	flow.CheckFatal(nat.StartGRPCServer())

	// Perform all network initialization so that DHCP client could
	// start sending packets
	flow.CheckFatal(flow.SystemInitPortsAndMemory())

	// Start DHCP client
	if nat.NeedDHCP || *setKniIP {
		nat.StartDHCPClient()
	}

	// Start flow scheduler
	go func() {
		flow.CheckFatal(flow.SystemStartScheduler())
	}()

	// Wait for interrupt
	sig := <-c
	fmt.Printf("Received signal %v\n", sig)
	nat.CloseAllDumpFiles()
}
