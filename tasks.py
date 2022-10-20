"""
Tasks for maintaining this project.
Run 'invoke --list' for guidance on using them
"""
import os

from invoke import task

PKG = "github.com/romnn/testcontainers"
CMD_PKG = PKG


ROOT_DIR = os.path.dirname(os.path.realpath(__file__))
BUILD_DIR = os.path.join(ROOT_DIR, "build")


@task
def format(c):
    """Format code"""
    c.run("pre-commit run go-fmt --all-files")
    c.run("pre-commit run go-imports --all-files")


@task
def embed(c):
    """Embed examples into README"""
    c.run("npx embedme README.md")


@task
def test(c):
    """Run tests"""
    # -coverpkg=all 
    c.run("env GO111MODULE=on go test -race -coverprofile=coverage.txt -covermode=atomic ./...")


@task
def cyclo(c):
    """Check code complexity"""
    c.run("pre-commit run go-cyclo --all-files")


@task
def lint(c):
    """Lint code"""
    c.run("pre-commit run go-lint --all-files")
    c.run("pre-commit run go-vet --all-files")


@task
def install_hooks(c):
    """Install pre-commit hooks"""
    c.run("pre-commit install")


@task
def pre_commit(c):
    """Run all pre-commit checks"""
    c.run("pre-commit run --all-files")


@task
def coverage(c):
    """Create coverage report"""
    cov_options = "-coverprofile=coverage.txt -coverpkg=all -covermode=atomic"
    c.run(f"env GO111MODULE=on go test -v -race {cov_options} ./...")


@task
def build(c):
    """Build the project"""
    c.run("pre-commit run go-build --all-files")


@task
def clean_build(c):
    """Clean up files from package building"""
    c.run(f"rm -fr {BUILD_DIR}")


@task
def clean_coverage(c):
    """Clean up files from coverage measurement"""
    c.run("find . -name 'coverage.txt' -exec rm -fr {} +")


@task(pre=[clean_build, clean_coverage])
def clean(c):
    """Runs all clean sub-tasks"""
    pass
