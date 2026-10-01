package deploycheck

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// k8sFiles returns every file under deploy/k8s, keyed by path relative to it.
func k8sFiles(t *testing.T) map[string]string {
	t.Helper()
	dir := filepath.Join(repoRoot(t), "deploy", "k8s")
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		out[rel] = read(t, p)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatal("no files found under deploy/k8s")
	}
	return out
}

var (
	secretKind  = regexp.MustCompile(`(?m)^kind:\s*Secret\s*$`)
	placeholder = regexp.MustCompile(`(?i)change[-_ ]?me`)
	pgURL       = regexp.MustCompile(`postgres(?:ql)?://[^\s"'` + "`" + `]+`)
	sslMode     = regexp.MustCompile(`sslmode=([a-z-]+)`)
	secretName  = regexp.MustCompile(`(?:secretRef|secretKeyRef):\s*\{\s*name:\s*([a-z0-9-]+)`)
	exampleVal  = regexp.MustCompile(`(?m)^\s{2}[A-Z_]+:\s*"(.*)"\s*$`)
)

// The kustomization must never deploy a Secret: credentials are created out of
// band, so a default (or placeholder) password can never reach a cluster.
func TestKustomizationContainsNoSecrets(t *testing.T) {
	files := k8sFiles(t)
	kust := files["kustomization.yaml"]
	if kust == "" {
		t.Fatal("missing kustomization.yaml")
	}
	for _, line := range strings.Split(kust, "\n") {
		name, ok := strings.CutPrefix(strings.TrimSpace(line), "- ")
		if !ok || !strings.HasSuffix(name, ".yaml") {
			continue
		}
		body, exists := files[name]
		if !exists {
			t.Errorf("kustomization lists %s, which does not exist", name)
			continue
		}
		if secretKind.MatchString(body) {
			t.Errorf("%s contains a Secret; create credentials out of band instead (deploy/k8s/README.md)", name)
		}
	}
	if strings.Contains(kust, "examples/") {
		t.Error("the secrets example must not be part of the kustomization")
	}
}

func TestNoPlaceholderCredentials(t *testing.T) {
	for name, body := range k8sFiles(t) {
		if placeholder.MatchString(body) {
			t.Errorf("%s contains a placeholder credential (CHANGE-ME and similar); use an explicitly invalid <set-via-secret-manager> value or a real out-of-band secret", name)
		}
	}
}

func TestDatabaseURLsRequireTLS(t *testing.T) {
	seen := 0
	for name, body := range k8sFiles(t) {
		for _, u := range pgURL.FindAllString(body, -1) {
			seen++
			m := sslMode.FindStringSubmatch(u)
			if m == nil || (m[1] != "require" && m[1] != "verify-ca" && m[1] != "verify-full") {
				t.Errorf("%s: PostgreSQL URL must use sslmode=require, verify-ca or verify-full: %s", name, u)
			}
		}
	}
	if seen == 0 {
		t.Error("expected at least one documented PostgreSQL URL")
	}
}

// sslmode=require needs a server with TLS; keep the bundled database in step.
func TestBundledPostgresEnablesTLS(t *testing.T) {
	pg := k8sFiles(t)["postgres.yaml"]
	for _, want := range []string{"ssl=on", "ssl_cert_file", "ssl_key_file"} {
		if !strings.Contains(pg, want) {
			t.Errorf("postgres.yaml must start PostgreSQL with %s so clients can use sslmode=require", want)
		}
	}
}

// The example exists only to show key names; every value must be unusable.
func TestSecretsExampleOnlyHasPlaceholders(t *testing.T) {
	body := k8sFiles(t)[filepath.Join("examples", "secrets.example.yaml")]
	if body == "" {
		t.Fatal("missing examples/secrets.example.yaml")
	}
	vals := exampleVal.FindAllStringSubmatch(body, -1)
	if len(vals) == 0 {
		t.Fatal("example defines no values")
	}
	for _, m := range vals {
		if !strings.HasPrefix(m[1], "<") || !strings.HasSuffix(m[1], ">") {
			t.Errorf("example value %q must be an explicit <placeholder>", m[1])
		}
	}
	if !strings.Contains(strings.ToLower(body), "not part of kustomization") && !strings.Contains(strings.ToLower(body), "not meant to be applied") {
		t.Error("example must state that it is a template and not to be applied")
	}
}

// Every secret a manifest needs must have a documented creation command.
func TestReferencedSecretsAreDocumented(t *testing.T) {
	files := k8sFiles(t)
	readme := files["README.md"]
	referenced := map[string]bool{}
	for name, body := range files {
		if strings.HasPrefix(name, "examples") || name == "README.md" {
			continue
		}
		for _, m := range secretName.FindAllStringSubmatch(body, -1) {
			referenced[m[1]] = true
		}
	}
	if len(referenced) == 0 {
		t.Fatal("expected the manifests to reference secrets")
	}
	for name := range referenced {
		if !strings.Contains(readme, "create secret generic "+name) {
			t.Errorf("secret %q is referenced but README.md has no `kubectl create secret generic %s` instructions", name, name)
		}
	}
	for _, key := range regexp.MustCompile(`key:\s*(POSTGRES_[A-Z]+)`).FindAllStringSubmatch(files["postgres.yaml"], -1) {
		if !strings.Contains(readme, "--from-literal="+key[1]+"=") {
			t.Errorf("README.md does not create the %s key that postgres.yaml reads", key[1])
		}
	}
}

func TestGitignoreBlocksSecretFiles(t *testing.T) {
	gi := read(t, filepath.Join(repoRoot(t), ".gitignore"))
	if !strings.Contains(gi, "*.secret.yaml") {
		t.Error(".gitignore should block *.secret.yaml so credentials are not committed by accident")
	}
}

// When kubectl is available, the rendered output must also be free of Secrets.
func TestRenderedManifestsHaveNoSecrets(t *testing.T) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl not installed")
	}
	out, err := exec.Command("kubectl", "kustomize", filepath.Join(repoRoot(t), "deploy", "k8s")).CombinedOutput()
	if err != nil {
		t.Fatalf("kubectl kustomize failed: %v\n%s", err, out)
	}
	if secretKind.Match(out) {
		t.Error("rendered manifests contain a Secret")
	}
	if !strings.Contains(string(out), "kind: StatefulSet") || !strings.Contains(string(out), "kind: Deployment") {
		t.Error("rendered output is missing the Deployment or StatefulSet")
	}
}
