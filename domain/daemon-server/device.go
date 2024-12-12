package daemon_server

// func EnsureAppRunning() error {
// 	p, err := NewProxy(flags.IsDev())
// 	if err != nil {
// 		return err
// 	}
//
// 	count := 0
// 	for {
// 		if p.Status() {
// 			return nil
// 		}
//
// 		if runtime.GOOS != "windows" {
// 			cmd := exec.Command("sudo", "echo", "")
// 			cmd.Stdin = os.Stdin
// 			cmd.Stderr = os.Stderr
// 			cmd.Stdout = os.Stdout
//
// 			err := cmd.Run()
// 			if err != nil {
// 				return err
// 			}
// 			command := exec.Command("sudo", flags.CliName, "app", "start")
// 			_ = command.Start()
//
// 		} else {
// 			_, err = functions.WinSudoExec(fmt.Sprintf("%s app start", flags.CliName), nil)
// 			if err != nil {
// 				functions.PrintError(err)
// 			}
// 		}
//
// 		count++
// 		if count >= 2 {
// 			return fmt.Errorf("failed to start app")
// 		}
//
// 		time.Sleep(2 * time.Second)
// 	}
// }
