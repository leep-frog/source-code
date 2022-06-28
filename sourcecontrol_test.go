package sourcecontrol

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/leep-frog/command"
)

func repoRunContents() []string {
	return []string{"set -e", "set -o pipefail", "git rev-parse --show-toplevel | xargs basename"}
}

func TestExecution(t *testing.T) {
	for _, test := range []struct {
		name string
		g    *git
		want *git
		etc  *command.ExecuteTestCase
	}{
		// TODO: Config tests
		// Simple command tests
		{
			name: "branch",
			etc: &command.ExecuteTestCase{
				Args: []string{"b"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{"git branch"},
				},
			},
		},
		{
			name: "pull",
			etc: &command.ExecuteTestCase{
				Args: []string{"l"},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						createSSHAgent,
						"git pull",
					},
				},
			},
		},
		// Checkout main
		{
			name: "checkout main",
			etc: &command.ExecuteTestCase{
				Args: []string{"m"},
				RunResponses: []*command.FakeRun{{
					Stdout: []string{"test-repo"},
				}},
				WantRunContents: [][]string{repoRunContents()},
				WantData: &command.Data{Values: map[string]interface{}{
					repoName.Name(): "test-repo",
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"git checkout main",
					},
				},
			},
		},
		{
			name: "checkout main uses default branch for unknown repo",
			g: &git{
				DefaultBranch: "mainer",
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"m"},
				RunResponses: []*command.FakeRun{{
					Stdout: []string{"test-repo"},
				}},
				WantRunContents: [][]string{repoRunContents()},
				WantData: &command.Data{Values: map[string]interface{}{
					repoName.Name(): "test-repo",
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"git checkout mainer",
					},
				},
			},
		},
		{
			name: "checkout main uses configured default branch for known repo",
			g: &git{
				DefaultBranch: "mainer",
				MainBranches: map[string]string{
					"test-repo": "mainest",
				},
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"m"},
				RunResponses: []*command.FakeRun{{
					Stdout: []string{"test-repo"},
				}},
				WantRunContents: [][]string{repoRunContents()},
				WantData: &command.Data{Values: map[string]interface{}{
					repoName.Name(): "test-repo",
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"git checkout mainest",
					},
				},
			},
		},
		// Merge main
		{
			name: "merge main",
			etc: &command.ExecuteTestCase{
				Args: []string{"mm"},
				RunResponses: []*command.FakeRun{{
					Stdout: []string{"test-repo"},
				}},
				WantRunContents: [][]string{repoRunContents()},
				WantData: &command.Data{Values: map[string]interface{}{
					repoName.Name(): "test-repo",
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"git merge main",
					},
				},
			},
		},
		{
			name: "merge main uses default branch for unknown repo",
			g: &git{
				DefaultBranch: "mainer",
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"mm"},
				RunResponses: []*command.FakeRun{{
					Stdout: []string{"test-repo"},
				}},
				WantRunContents: [][]string{repoRunContents()},
				WantData: &command.Data{Values: map[string]interface{}{
					repoName.Name(): "test-repo",
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"git merge mainer",
					},
				},
			},
		},
		{
			name: "merge main uses configured default branch for known repo",
			g: &git{
				DefaultBranch: "mainer",
				MainBranches: map[string]string{
					"test-repo": "mainest",
				},
			},
			etc: &command.ExecuteTestCase{
				Args: []string{"mm"},
				RunResponses: []*command.FakeRun{{
					Stdout: []string{"test-repo"},
				}},
				WantRunContents: [][]string{repoRunContents()},
				WantData: &command.Data{Values: map[string]interface{}{
					repoName.Name(): "test-repo",
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						"git merge mainest",
					},
				},
			},
		},
		// Commit
		{
			name: "commit requires args",
			etc: &command.ExecuteTestCase{
				Args:       []string{"c"},
				WantStderr: []string{`Argument "MESSAGE" requires at least 1 argument, got 0`},
				WantErr:    fmt.Errorf(`Argument "MESSAGE" requires at least 1 argument, got 0`),
			},
		},
		{
			name: "simple commit",
			etc: &command.ExecuteTestCase{
				Args: []string{"c", "did", "things"},
				WantData: &command.Data{Values: map[string]interface{}{
					messageArg.Name(): []string{"did", "things"},
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						`git commit -m  "did things"`,
						"echo Success!",
					},
				},
			},
		},
		{
			name: "commit no verify",
			etc: &command.ExecuteTestCase{
				Args: []string{"c", "did", "things", "-n"},
				WantData: &command.Data{Values: map[string]interface{}{
					messageArg.Name(): []string{"did", "things"},
					nvFlag.Name():     true,
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						`git commit -m --no-verify "did things"`,
						"echo Success!",
					},
				},
			},
		},
		{
			name: "commit push",
			etc: &command.ExecuteTestCase{
				Args: []string{"c", "did", "things", "-p"},
				WantData: &command.Data{Values: map[string]interface{}{
					messageArg.Name(): []string{"did", "things"},
					pushFlag.Name():   true,
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						`git commit -m  "did things"`,
						createSSHAgent,
						`git push`,
						"echo Success!",
					},
				},
			},
		},
		{
			name: "commit no verify and push",
			etc: &command.ExecuteTestCase{
				Args: []string{"c", "did", "things", "--no-verify", "--push"},
				WantData: &command.Data{Values: map[string]interface{}{
					messageArg.Name(): []string{"did", "things"},
					nvFlag.Name():     true,
					pushFlag.Name():   true,
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						`git commit -m --no-verify "did things"`,
						createSSHAgent,
						`git push`,
						"echo Success!",
					},
				},
			},
		},
		// Commit & push
		{
			name: "commit and push requires args",
			etc: &command.ExecuteTestCase{
				Args:       []string{"cp"},
				WantStderr: []string{`Argument "MESSAGE" requires at least 1 argument, got 0`},
				WantErr:    fmt.Errorf(`Argument "MESSAGE" requires at least 1 argument, got 0`),
			},
		},
		{
			name: "simple commit and push",
			etc: &command.ExecuteTestCase{
				Args: []string{"cp", "did", "things"},
				WantData: &command.Data{Values: map[string]interface{}{
					messageArg.Name(): []string{"did", "things"},
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						`git commit -m  "did things"`,
						createSSHAgent,
						`git push`,
						"echo Success!",
					},
				},
			},
		},
		{
			name: "commit and push no verify",
			etc: &command.ExecuteTestCase{
				Args: []string{"cp", "did", "things", "-n"},
				WantData: &command.Data{Values: map[string]interface{}{
					messageArg.Name(): []string{"did", "things"},
					nvFlag.Name():     true,
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						`git commit -m --no-verify "did things"`,
						createSSHAgent,
						`git push`,
						"echo Success!",
					},
				},
			},
		},
		// Checkout new branch
		{
			name: "checkout branch requires arg",
			etc: &command.ExecuteTestCase{
				Args:       []string{"ch"},
				WantStderr: []string{`Argument "BRANCH" requires at least 1 argument, got 0`},
				WantErr:    fmt.Errorf(`Argument "BRANCH" requires at least 1 argument, got 0`),
			},
		},
		{
			name: "checkout branch requires one arg",
			etc: &command.ExecuteTestCase{
				Args: []string{"ch", "tree", "limb"},
				WantData: &command.Data{Values: map[string]interface{}{
					branchArg.Name(): "tree",
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						`git checkout tree`,
					},
				},
				WantStderr: []string{`Unprocessed extra args: [limb]`},
				WantErr:    fmt.Errorf(`Unprocessed extra args: [limb]`),
			},
		},
		{
			name: "checks out a branch",
			etc: &command.ExecuteTestCase{
				Args: []string{"ch", "tree"},
				WantData: &command.Data{Values: map[string]interface{}{
					branchArg.Name(): "tree",
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						`git checkout tree`,
					},
				},
			},
		},
		{
			name: "checks out a new branch",
			etc: &command.ExecuteTestCase{
				Args: []string{"ch", "tree", "-n"},
				WantData: &command.Data{Values: map[string]interface{}{
					branchArg.Name():     "tree",
					newBranchFlag.Name(): true,
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						`git checkout -b tree`,
					},
				},
			},
		},
		// Delete new branch
		{
			name: "delete branch requires arg",
			etc: &command.ExecuteTestCase{
				Args:       []string{"bd"},
				WantStderr: []string{`Argument "BRANCH" requires at least 1 argument, got 0`},
				WantErr:    fmt.Errorf(`Argument "BRANCH" requires at least 1 argument, got 0`),
			},
		},
		{
			name: "delete branch requires one arg",
			etc: &command.ExecuteTestCase{
				Args: []string{"bd", "tree", "limb"},
				WantData: &command.Data{Values: map[string]interface{}{
					branchArg.Name(): "tree",
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						`git branch -d tree`,
					},
				},
				WantStderr: []string{`Unprocessed extra args: [limb]`},
				WantErr:    fmt.Errorf(`Unprocessed extra args: [limb]`),
			},
		},
		{
			name: "deletes a branch",
			etc: &command.ExecuteTestCase{
				Args: []string{"bd", "tree"},
				WantData: &command.Data{Values: map[string]interface{}{
					branchArg.Name(): "tree",
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						`git branch -d tree`,
					},
				},
			},
		},
		{
			name: "force deletes a branch",
			etc: &command.ExecuteTestCase{
				Args: []string{"bd", "-f", "tree"},
				WantData: &command.Data{Values: map[string]interface{}{
					branchArg.Name():   "tree",
					forceDelete.Name(): true,
				}},
				WantExecuteData: &command.ExecuteData{
					Executable: []string{
						`git branch -D tree`,
					},
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.g == nil {
				test.g = CLI()
			}
			test.etc.Node = test.g.Node()
			command.ExecuteTest(t, test.etc)
			command.ChangeTest(t, test.want, test.g, cmpopts.IgnoreUnexported(git{}), cmpopts.EquateEmpty())
		})
	}
}
