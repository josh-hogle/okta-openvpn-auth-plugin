package main

import (
	"fmt"
	"io/ioutil"
	golog "log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/Masterminds/semver"
	tberrors "github.com/josh-hogle/go-toolbox/errors"
	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/app"
	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/errors"
	"github.com/josh-hogle/zerolog/v2"
	"github.com/josh-hogle/zerolog/v2/log"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// initialize logging
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000Z07:00"
	stdoutConsoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	stderrConsoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
	stdoutLevelWriter := zerolog.NewFilteredLevelWriter([]zerolog.Level{
		zerolog.DebugLevel, zerolog.InfoLevel, zerolog.WarnLevel,
	}, stdoutConsoleWriter)
	stderrLevelWriter := zerolog.NewFilteredLevelWriter([]zerolog.Level{
		zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel,
	}, stderrConsoleWriter)
	writer := zerolog.MultiLevelWriter(stdoutLevelWriter, stderrLevelWriter)
	l := zerolog.New(writer).With().Timestamp().Logger()
	l.SetLevel(zerolog.InfoLevel)
	log.ReplaceGlobal(l)
	golog.SetOutput(ioutil.Discard) // discard standard logger output

	// initialize application settings
	if app.DevBuildStr == "" {
		app.DevBuildStr = "false"
	}
	flag, err := strconv.ParseBool(app.DevBuildStr)
	if err != nil { // should never happen as we control this through the build process
		e := &errors.GeneralFailure{
			Err: err,
			Msg: fmt.Sprintf("DevBuild is not a valid boolean value: %s", err.Error()),
		}
		log.Error().Err(e.InternalError()).Msg(e.Error())
		os.Exit(e.Code())
	}
	app.DevBuild = flag

	// convert app version to a semantic version
	semVer, err := semver.NewVersion(app.Version)
	if err != nil { // should never happen as we control this through the build process
		e := &errors.GeneralFailure{
			Err: err,
			Msg: fmt.Sprintf("Version is not a valid semantic version: %s", err.Error()),
		}
		log.Error().Err(e.InternalError()).Msg(e.Error())
		os.Exit(e.Code())
	}
	app.SemanticVersion = semVer

	// execute the command
	var exitCode int
	err = NewRootCommand().Execute()
	if e, ok := err.(tberrors.ExtendedError); ok {
		// the extended error message should already have been logged during execution
		exitCode = e.Code()
	} else if err != nil {
		// error returned was not an "extended" error so treat it as a usage error
		e := &errors.Usage{
			Err: err,
		}
		log.Error().Err(e.InternalError()).Msg(e.Error())
		exitCode = e.Code()
	}
	if exitCode != 0 {
		log.Info().Int("exit_code", exitCode).Msgf("exiting with non-zero exit code: %d", exitCode)
	}
	os.Exit(exitCode)
}
