package main

import (
    "fmt"
    "os/exec"
    "regexp"
    "strconv"

    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

func scanForIDRAC() (string, error) {
    fmt.Println("Starting network scan for iDRAC...")
    cmd := exec.Command("nmap", "-p", "443", "--open", "192.168.1.0/24")
    output, err := cmd.Output()
    if err != nil {
        fmt.Println("Error executing nmap:", err)
        return "", err
    }
    fmt.Println("nmap output:", string(output))

    re := regexp.MustCompile(`Nmap scan report for idrac-[\w-]+ \(([\d\.]+)\)`)
    matches := re.FindStringSubmatch(string(output))

    if len(matches) == 2 {
        return matches[1], nil
    }
    return "", fmt.Errorf("idrac not found")
}

func executeIPMICommand(ip, user, pass, command string) string {
    fmt.Println("Executing IPMI command:", command)
    fullCommand := fmt.Sprintf("ipmitool -I lanplus -H %s -U %s -P \"%s\" %s", ip, user, pass, command)
    out, err := exec.Command("sh", "-c", fullCommand).Output()
    if err != nil {
        fmt.Println("Error executing IPMI command:", err)
        return err.Error()
    }
    return string(out)
}

func main() {
    fmt.Println("Starting application...")
    a := app.New()
    w := a.NewWindow("Server Management")

    ipEntry := widget.NewEntry()
    ipEntry.SetPlaceHolder("IP Address (Default to local IP network)")

    userEntry := widget.NewEntry()
    userEntry.SetPlaceHolder("Username")

    passEntry := widget.NewPasswordEntry()
    passEntry.SetPlaceHolder("Password")

    resultLabel := widget.NewLabel("")

    scanButton := widget.NewButton("Scan for iDRAC", func() {
        ip, err := scanForIDRAC()
        if err != nil {
            resultLabel.SetText("Error scanning for iDRAC: " + err.Error())
        } else {
            ipEntry.SetText(ip)
            resultLabel.SetText("Found iDRAC at " + ip)
        }
    })

    powerOnBtn := widget.NewButton("Power On", func() {
        result := executeIPMICommand(ipEntry.Text, userEntry.Text, passEntry.Text, "chassis power on")
        resultLabel.SetText(result)
    })

    powerOffBtn := widget.NewButton("Power Off", func() {
        result := executeIPMICommand(ipEntry.Text, userEntry.Text, passEntry.Text, "chassis power off")
        resultLabel.SetText(result)
    })

    fanSpeedEntry := widget.NewEntry()
    fanSpeedEntry.SetPlaceHolder("Fan Speed %")

    setFanSpeedBtn := widget.NewButton("Set Fan Speed", func() {
        speed, err := strconv.Atoi(fanSpeedEntry.Text)
        if err == nil && speed >= 0 && speed <= 100 {
            command := fmt.Sprintf("raw 0x30 0x30 0x02 0xff %x", speed)
            result := executeIPMICommand(ipEntry.Text, userEntry.Text, passEntry.Text, command)
            resultLabel.SetText(result)
        } else {
            resultLabel.SetText("Invalid speed value")
        }
    })

    enableDynamicFanControlChk := widget.NewCheck("Enable dynamic fan control", func(checked bool) {
        var result string
        if checked {
            result = executeIPMICommand(ipEntry.Text, userEntry.Text, passEntry.Text, "raw 0x30 0x30 0x01 0x01")
        } else {
            result = executeIPMICommand(ipEntry.Text, userEntry.Text, passEntry.Text, "raw 0x30 0x30 0x01 0x00")
        }
        resultLabel.SetText(result)
    })

    form := container.NewVBox(
        ipEntry,
        userEntry,
        passEntry,
        scanButton,
        powerOnBtn,
        powerOffBtn,
        fanSpeedEntry,
        setFanSpeedBtn,
        enableDynamicFanControlChk,
        resultLabel,
    )

    w.SetContent(form)
    w.ShowAndRun()

    fmt.Println("Application running...")
}
