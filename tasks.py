from invoke import task
from pathlib import Path

PKG = "github.com/romnn/testcontainers"
CMD_PKG = PKG

ROOT_DIR = Path(__file__).parent
BUILD_DIR = ROOT_DIR / "build"


@task
def format(c):
    """Format code"""
    c.run("pre-commit run go-fmt --all-files")
    c.run("pre-commit run go-imports --all-files")


@task
def embed(c):
    """Embed examples into README"""
    c.run(f"npx embedme {ROOT_DIR / 'README.md'}")


@task
def test(c):
    """Run tests"""
    cmd = [
        "go",
        "test",
        "-race",
        "-coverpkg=all",
        "-coverprofile=coverage.txt",
        "-covermode=atomic",
        "./...",
    ]
    c.run(" ".join(cmd))


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
