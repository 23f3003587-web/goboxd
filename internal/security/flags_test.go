package security

import (
	"strings"
	"testing"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/model"
)

func TestValidateBuildFlags(t *testing.T) {
	if err := config.LoadLanguages("../../languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}

	if err := ValidateBuildFlags("cpp", []string{"-O2", "-std=c++17"}); err != nil {
		t.Fatalf("expected valid build flags, got error: %v", err)
	}

	if err := ValidateBuildFlags("cpp", []string{"-g"}); err == nil {
		t.Fatal("expected disallowed build flag to return an error")
	}
}

func TestValidateRunFlags(t *testing.T) {
	if err := config.LoadLanguages("../../languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}

	if err := ValidateRunFlags("py3", []string{"-u"}); err == nil {
		t.Fatal("expected run flags to be rejected when no allowlist is configured")
	}
}
func TestValidateRunRequest(t *testing.T) {
	if err := config.LoadLanguages("../../languages.yaml"); err != nil {
		t.Fatalf("failed to load languages: %v", err)
	}

	validReq := model.RunRequest{
		Language: "py3",
		Source:   "print(1)",
		Tests:    []model.TestCase{{Stdin: "", ExpectedStdout: "1"}},
	}
	if err := ValidateRunRequest(validReq); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}

	invalidFilenameReq := validReq
	invalidFilenameReq.SourceFilename = "../../etc/passwd"
	if err := ValidateRunRequest(invalidFilenameReq); err == nil {
		t.Fatal("expected request with invalid source filename to be rejected")
	}

	invalidBuildFlagsReq := validReq
	invalidBuildFlagsReq.Language = "cpp"
	invalidBuildFlagsReq.Source = "#include <iostream>\nint main(){std::cout<<\"hi\";return 0;}"
	invalidBuildFlagsReq.Build = &model.BuildRun{Flags: []string{"-O999"}}
	if err := ValidateRunRequest(invalidBuildFlagsReq); err == nil {
		t.Fatal("expected request with disallowed build flags to be rejected")
	}

	invalidRunFlagsReq := validReq
	invalidRunFlagsReq.Run = &model.BuildRun{Flags: []string{"-u"}}
	if err := ValidateRunRequest(invalidRunFlagsReq); err == nil {
		t.Fatal("expected request with disallowed run flags to be rejected")
	}

	tooLargeStdinReq := validReq
	tooLargeStdinReq.Tests = []model.TestCase{{Stdin: strings.Repeat("a", config.Global.MaxTestStdinBytes+1), ExpectedStdout: "1"}}
	if err := ValidateRunRequest(tooLargeStdinReq); err == nil {
		t.Fatal("expected request with too-large stdin to be rejected")
	}

	tooLargeExpectedReq := validReq
	tooLargeExpectedReq.Tests = []model.TestCase{{Stdin: "", ExpectedStdout: strings.Repeat("a", config.Global.MaxExpectedStdoutBytes+1)}}
	if err := ValidateRunRequest(tooLargeExpectedReq); err == nil {
		t.Fatal("expected request with too-large expected_stdout to be rejected")
	}
}
