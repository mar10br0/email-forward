package log

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type ANSI_CODES = struct {
	RED    string
	GREEN  string
	YELLOW string
	BLUE   string
	PURPLE string
	CYAN   string
	WHITE  string
	RESET  string
}

var ANSI_MONO = ANSI_CODES{
	RED:    "",
	GREEN:  "",
	YELLOW: "",
	BLUE:   "",
	PURPLE: "",
	CYAN:   "",
	WHITE:  "",
	RESET:  "",
}
var ANSI_COLOR = ANSI_CODES{
	RED:    "\033[31m",
	GREEN:  "\033[32m",
	YELLOW: "\033[33m",
	BLUE:   "\033[34m",
	PURPLE: "\033[35m",
	CYAN:   "\033[36m",
	WHITE:  "\033[37m",
	RESET:  "\033[0m",
}

var ANSI = ANSI_MONO
var TIMED = false
var VERBOSE = false
var TIME_FORMAT = "2006/01/02 15:04:05"

var LOGLEVEL = map[string]aws.LogLevelType{
	"":                            aws.LogOff,
	"LogOff":                      aws.LogOff,
	"LogDebug":                    aws.LogDebug,
	"LogDebugWithSigning":         aws.LogDebugWithSigning,
	"LogDebugWithHTTPBody":        aws.LogDebugWithHTTPBody,
	"LogDebugWithRequestRetries":  aws.LogDebugWithRequestRetries,
	"LogDebugWithRequestErrors":   aws.LogDebugWithRequestErrors,
	"LogDebugWithEventStreamBody": aws.LogDebugWithEventStreamBody,
	"LogDebugWithDeprecated":      aws.LogDebugWithDeprecated,
}[os.Getenv("LOGLEVEL")]
var IS_DESKTOP = os.Getenv("LAMBDA_TASK_ROOT") == ""

func init() {
	if IS_DESKTOP {
		ANSI = ANSI_COLOR
	} else {
		ANSI = ANSI_MONO
	}
	TIMED = IS_DESKTOP
	VERBOSE = LOGLEVEL.AtLeast(aws.LogDebug)
}

func desktopTime() string {
	if TIMED {
		return time.Now().Format(TIME_FORMAT)
	}
	return ""
}

func Progress(format string, args ...interface{}) {
	fmt.Printf("%s%s %s%s\n", ANSI.YELLOW, desktopTime(), fmt.Sprintf(format, args...), ANSI.RESET)
}

func Debug(format string, args ...interface{}) {
	if VERBOSE {
		fmt.Printf("%s%s DEBUG: %s%s\n", ANSI.BLUE, desktopTime(), fmt.Sprintf(format, args...), ANSI.RESET)
	}
}

func Response(response http.Header) {
	if VERBOSE {
		fmt.Printf("%s---[ RESPONSE ]--------------------------------------\n", ANSI.BLUE)
		for key, values := range response {
			fmt.Printf("%s: %s\n", key, strings.Join(values, ", "))
		}
		fmt.Printf("-----------------------------------------------------%s\n", ANSI.RESET)
	}
}

func Failure(err error, format string, args ...interface{}) {
	if format != "" {
		fmt.Printf("%s%s ERROR: %s\n  %s%s\n", ANSI.RED, desktopTime(), fmt.Sprintf(format, args...), err.Error(), ANSI.RESET)
	} else {
		fmt.Printf("%s%s ERROR: %s%s\n", ANSI.RED, desktopTime(), err.Error(), ANSI.RESET)
	}
}

func Success(format string, args ...interface{}) {
	fmt.Printf("%s%s SUCCESS: %s%s\n", ANSI.GREEN, desktopTime(), fmt.Sprintf(format, args...), ANSI.RESET)
}

func Plurality(count int64, plural string) string {
	if count == 1 {
		return fmt.Sprintf("1 %s", strings.TrimSuffix(plural, "s"))
	}
	var p *message.Printer = message.NewPrinter(language.English)
	return p.Sprintf("%v %s", count, plural)
}
