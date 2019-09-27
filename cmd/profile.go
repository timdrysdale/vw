package cmd

/*

	// handle memprofile
	go func() {

		if memprofile != "" {

			time.Sleep(time.Duration(duration) * time.Second)

			f, err := os.Create(memprofile)

			if err != nil {
				log.WithField("error", err).Fatal("Could not create memory profile")
			}

			defer f.Close()

			if err := pprof.WriteHeapProfile(f); err != nil {
				log.WithField("error", err).Fatal("Could not write memory profile")
			}

			defer pprof.StopCPUProfile()
			close(closed)
		}
	}()

	// handle cpuprofile
	if cpuprofile != "" {

		f, err := os.Create(cpuprofile)

		if err != nil {
			log.WithField("error", err).Fatal("Could not create CPU profile")
		}

		defer f.Close()

		if err := pprof.StartCPUProfile(f); err != nil {
			log.WithField("error", err).Fatal("Could not start CPU profile")
		}

		defer pprof.StopCPUProfile()

	}
*/
