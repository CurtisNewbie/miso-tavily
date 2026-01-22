package tavily

import (
	"github.com/curtisnewbie/miso/errs"
	"github.com/curtisnewbie/miso/miso"
	"github.com/curtisnewbie/miso/util/json"
	"github.com/curtisnewbie/miso/util/strutil"
	"github.com/tmaxmax/go-sse"
)

var (
	ResearchURL = "https://api.tavily.com/research"
)

type InitResearchReq struct {
	CitationFormat string        `json:"citation_format"` // numbered, mla, apa, chicago
	Input          string        `json:"input"`
	Model          string        `json:"model"` // mini, pro, auto
	OutputSchema   *OutputSchema `json:"output_schema"`
	Stream         bool          `json:"stream"`
}

type OutputSchema struct {
	Properties struct {
		Company struct {
			Description string `json:"description"`
			Type        string `json:"type"`
		} `json:"company"`
		FinancialDetails struct {
			Description string `json:"description"`
			Properties  struct {
				OperatingIncome struct {
					Description string `json:"description"`
					Type        string `json:"type"`
				} `json:"operating_income"`
			} `json:"properties"`
			Type string `json:"type"`
		} `json:"financial_details"`
		KeyMetrics struct {
			Description string `json:"description"`
			Items       struct {
				Type string `json:"type"`
			} `json:"items"`
			Type string `json:"type"`
		} `json:"key_metrics"`
	} `json:"properties"`
	Required []string `json:"required"`
}

type Choice struct {
	Delta Delta `json:"delta"`
}

type Delta struct {
	Content   string `json:"content"`
	Role      string `json:"role"`
	ToolCalls *struct {
		ToolCall []ToolCall `json:"tool_call"`
		Type     string     `json:"type"`
	} `json:"tool_calls"`
	Sources []Source `json:"sources"`
}

type Source struct {
	Favicon string `json:"favicon"`
	Title   string `json:"title"`
	URL     string `json:"url"`
}

type ToolCall struct {
	Arguments string   `json:"arguments"`
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Queries   []string `json:"queries"`
}

// https://docs.tavily.com/documentation/api-reference/endpoint/research-streaming
type StreamResearchEvent struct {
	Choices []Choice `json:"choices"`
	Created int64    `json:"created"`
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Object  string   `json:"object"`
}

type ResearchProgress struct {
	Name      string
	Arguments string
	Queries   []string
}

type ProgressHook func(ResearchProgress) error
type SourceHook func([]Source) error

type streamResearchOp struct {
	ProgressHook ProgressHook
	SourceHook   SourceHook
}

func WithProgressHook(p ProgressHook) StreamResearchOpFunc {
	return func(o *streamResearchOp) {
		o.ProgressHook = p
	}
}

func WithSourceHook(p SourceHook) StreamResearchOpFunc {
	return func(o *streamResearchOp) {
		o.SourceHook = p
	}
}

type StreamResearchOpFunc func(o *streamResearchOp)

func StreamResearch(rail miso.Rail, apiKey string, req InitResearchReq, ops ...StreamResearchOpFunc) (string, error) {
	if req.Input == "" {
		return "", errs.NewErrf("Input required")
	}
	if req.CitationFormat == "" {
		req.CitationFormat = "numbered"
	}
	if req.Model == "" {
		req.Model = "mini"
	}

	sro := &streamResearchOp{}
	for _, op := range ops {
		if op == nil {
			continue
		}
		op(sro)
	}
	req.Stream = true
	body := strutil.NewBuilder()
	err := miso.NewClient(rail, ResearchURL).
		AddAuthBearer(apiKey).
		Require2xx().
		PostJson(req).
		Sse(func(e sse.Event) (stop bool, err error) {
			if miso.IsShuttingDown() {
				return true, miso.ErrServerShuttingDown.New()
			}
			if e.Data == "" {
				return false, nil
			}
			switch e.Type {
			case "done":
				if sro.ProgressHook != nil {
					if err := sro.ProgressHook(ResearchProgress{
						Name:      "Done",
						Arguments: "Research Completed",
					}); err != nil {
						return true, err
					}
				}
				return true, nil
			case "sources":
				sre, err := json.SParseJsonAs[StreamResearchEvent](e.Data)
				if err != nil {
					return false, err
				}

				if sro.SourceHook != nil {
					n := 0
					for _, c := range sre.Choices {
						n += len(c.Delta.Sources)
					}
					s := make([]Source, 0, n)
					for _, c := range sre.Choices {
						for _, v := range c.Delta.Sources {
							s = append(s, v)
						}
					}
					if err := sro.SourceHook(s); err != nil {
						return true, err
					}
				}
			default:
				sre, err := json.SParseJsonAs[StreamResearchEvent](e.Data)
				if err != nil {
					return false, err
				}

				for _, c := range sre.Choices {
					if c.Delta.Content != "" {
						body.WriteString(c.Delta.Content)
					} else if c.Delta.ToolCalls != nil {
						if c.Delta.ToolCalls.Type == "tool_call" {
							for _, tc := range c.Delta.ToolCalls.ToolCall {
								rail.Infof("[%v] %v", tc.Name, tc.Arguments)
								if sro.ProgressHook != nil {
									if err := sro.ProgressHook(ResearchProgress{
										Name:      tc.Name,
										Arguments: tc.Arguments,
										Queries:   tc.Queries,
									}); err != nil {
										return true, err
									}
								}
							}
						}
					}
				}
			}
			return false, nil
		})
	if err != nil {
		return "", err
	}
	return body.String(), nil
}
