package tavily

import (
	"testing"

	"github.com/curtisnewbie/miso/flow"
)

func TestStreamResearch(t *testing.T) {
	rail := flow.EmptyRail()
	reportContent, err := StreamResearch(rail, "my-key", InitResearchReq{
		CitationFormat: "numbered",
		Input:          "How finetuning works?",
		Model:          "mini",
	}, WithProgressHook(func(rp ResearchProgress) error {
		rail.Infof("Progress: %v - %v", rp.Name, rp.Arguments)
	}))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(reportContent)
}
