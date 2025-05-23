package retro

import (
	"fmt"
	"os"
	"time"
)

func ShowSplash() {
	showPHEAR()
	time.Sleep(1000 * time.Millisecond)
//	showXCRAFT()
//	time.Sleep(800 * time.Millisecond)
	fmt.Fprintln(os.Stdout, "\n   Customize your apps, but cooler 🕹️")
}

func showPHEAR() {
	phear := `
██████╗ ██╗  ██╗███████╗ █████╗ ██████╗ 
██╔══██╗██║  ██║██╔════╝██╔══██╗██╔══██╗
██████╔╝███████║█████╗  ███████║██████╔╝
██╔═══╝ ██╔══██║██╔══╝  ██╔══██║██╔══██╗
██║     ██║  ██║██      ██║  ██║██║  ██║
██║     ██║  ██║███████╗██║  ██║██║  ██║
╚═╝     ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝
`
	fmt.Fprintln(os.Stdout, phear)
}

func showXCRAFT() {
	xcraft := `
██╗  ██╗ ██████╗██████╗  █████╗ ███████╗████████╗
╚██╗██╔╝██╔════╝██╔═ ██╗██╔══██╗██╔════╝╚══██╔══╝
 ╚███╔╝ ██║     ███████╔╝███████║█████╗     ██║   
 ██╔██╗ ██║     ██╔══██╗██╔══██║██╔══╝     ██║   
██╔╝ ██╗╚██████╗██║  ██║██║  ██║██         ██║   
██╔╝ ██╗╚██████╗██║  ██║██║  ██║██         ██║   
╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═         ╚═╝   
`
	fmt.Fprintln(os.Stdout, xcraft)
}