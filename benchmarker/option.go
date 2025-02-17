package benchmarker

import (
	"fmt"
	"risuwork-benchmarker/scenario"
	"strings"
	"time"
)

// Option is for benchmarker settings
type Option struct {
	scenario.Option
	ExitErrorOnFail bool
	PrepareOnly     bool
	LoadTimeout     time.Duration
}

// fmt.Stringer インターフェースを実装
// log.Print などに渡した際、このメソッドが実装されていれば返した文字列が出力される
func (o Option) String() string {
	args := []string{
		"benchmarker",
		fmt.Sprintf("--target-host=%s", o.TargetHost),
		fmt.Sprintf("--request-timeout=%s", o.RequestTimeout.String()),
		fmt.Sprintf("--prepare-request-timeout=%s", o.PrepareRequestTimeout.String()),
		fmt.Sprintf("--initialize-request-timeout=%s", o.InitializeRequestTimeout.String()),
		fmt.Sprintf("--exit-error-on-fail=%v", o.ExitErrorOnFail),
		fmt.Sprintf("--prepare-only=%v", o.PrepareOnly),
		fmt.Sprintf("--load-interval=%v", o.LoadTimeout),
	}

	return strings.Join(args, " ")
}
