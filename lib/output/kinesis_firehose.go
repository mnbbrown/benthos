// Copyright (c) 2018 Ashley Jeffs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package output

import (
	"fmt"

	"github.com/Jeffail/benthos/v3/lib/log"
	"github.com/Jeffail/benthos/v3/lib/message/batch"
	"github.com/Jeffail/benthos/v3/lib/metrics"
	"github.com/Jeffail/benthos/v3/lib/output/writer"
	"github.com/Jeffail/benthos/v3/lib/types"
)

//------------------------------------------------------------------------------

func init() {
	Constructors[TypeKinesisFirehose] = TypeSpec{
		constructor: NewKinesisFirehose,
		Description: `
Sends messages to a Kinesis Firehose delivery stream.

### Credentials

By default Benthos will use a shared credentials file when connecting to AWS
services. It's also possible to set them explicitly at the component level,
allowing you to transfer data across accounts. You can find out more
[in this document](../aws.md).`,
		sanitiseConfigFunc: func(conf Config) (interface{}, error) {
			return sanitiseWithBatch(conf.KinesisFirehose, conf.KinesisFirehose.Batching)
		},
		Async:   true,
		Batches: true,
	}
}

//------------------------------------------------------------------------------

// NewKinesisFirehose creates a new KinesisFirehose output type.
func NewKinesisFirehose(conf Config, mgr types.Manager, log log.Modular, stats metrics.Type) (Type, error) {
	kin, err := writer.NewKinesisFirehose(conf.KinesisFirehose, log, stats)
	if err != nil {
		return nil, err
	}
	var w Type
	if conf.KinesisFirehose.MaxInFlight == 1 {
		w, err = NewWriter(
			TypeKinesisFirehose, kin, log, stats,
		)
	} else {
		w, err = NewAsyncWriter(
			TypeKinesisFirehose, conf.KinesisFirehose.MaxInFlight, kin, log, stats,
		)
	}
	if bconf := conf.KinesisFirehose.Batching; err == nil && !bconf.IsNoop() {
		policy, err := batch.NewPolicy(bconf, mgr, log.NewModule(".batching"), metrics.Namespaced(stats, "batching"))
		if err != nil {
			return nil, fmt.Errorf("failed to construct batch policy: %v", err)
		}
		w = NewBatcher(policy, w, log, stats)
	}
	return w, err
}

//------------------------------------------------------------------------------
