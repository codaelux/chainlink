package pipeline

import (
	"net/url"

	// "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pkg/errors"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/store/models"
)

type BridgeTask struct {
	BaseTask `mapstructure:",squash"`

	Name        string          `json:"name"`
	RequestData HttpRequestData `json:"requestData"`

	orm    ORM
	config Config
}

var _ Task = (*BridgeTask)(nil)

func (t *BridgeTask) Type() TaskType {
	return TaskTypeBridge
}

func (t *BridgeTask) Run(inputs []Result) (result Result) {
	if len(inputs) > 0 {
		return Result{Error: errors.Wrapf(ErrWrongInputCardinality, "BridgeTask requires 0 inputs")}
	}

	url, err := t.getBridgeURLFromName()
	if err != nil {
		return Result{Error: err}
	}

	// client := &http.Client{Timeout: t.config.DefaultHTTPTimeout().Duration(), Transport: http.DefaultTransport}
	// client.Transport = promhttp.InstrumentRoundTripperDuration(promFMResponseTime, client.Transport)
	// client.Transport = instrumentRoundTripperReponseSize(promFMResponseSize, client.Transport)

	// add an arbitrary "id" field to the request json
	// this is done in order to keep request payloads consistent in format
	// between flux monitor polling requests and http/bridge adapters
	if t.RequestData == nil {
		t.RequestData = HttpRequestData{}
	}
	t.RequestData["id"] = models.NewID()

	result = (&HTTPTask{
		URL:                            models.WebURL(url),
		Method:                         "POST",
		RequestData:                    t.RequestData,
		AllowUnrestrictedNetworkAccess: true,
		config:                         t.config,
	}).Run(inputs)
	if result.Error != nil {
		return result
	}
	logger.Debugw("Bridge: fetched answer",
		"answer", string(result.Value.([]byte)),
		"url", url.String(),
	)
	return result
}

func (t BridgeTask) getBridgeURLFromName() (url.URL, error) {
	task := models.TaskType(t.Name)
	bridge, err := t.orm.FindBridge(task)
	if err != nil {
		return url.URL{}, err
	}
	bridgeURL := url.URL(bridge.URL)
	return bridgeURL, nil
}
