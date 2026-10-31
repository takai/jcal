# jcal

jcal is a command-line calendar program that displays Japanese national holidays.

## Directory Structure

* Follow the Standard Go Project Layout.

## Specifications

* See `docs/spec.md` for details. Update this file whenever you change system behavior.

## Build Tools

* `mise build` – Build the project
* `mise format` – Apply Go fixes and format the code
* `mise test` – Run tests

## Development Guidelines

* Use the red/green/refactor TDD cycle
* After making code changes, run `mise format` and `mise test`
* Use `git ai-commit` to create commits.
  * Always pass a short human-readable summary in English with the `--context` option when running `git ai-commit`.
