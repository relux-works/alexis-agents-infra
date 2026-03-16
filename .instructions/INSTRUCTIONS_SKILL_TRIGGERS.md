# Skill Triggers

Automatic skill activation rules. When these topics come up — **FIRST load the skill via Skill tool**, then proceed.

## Language Policy

- **Specs, docs, code comments:** English only
- **Sub-agent prompts:** English only
- **Task/story/epic descriptions:** English only
- **User communication:** Match user's language (default: English)

---

## Triggers

| Triggers | Skill | Action |
|----------|-------|--------|
| architecture diagram, C4, structurizr, plantuml, sequence/component/class diagram | `architecture-diagrams` | Follow skill for patterns, folder structure, rendering. |
| iOS UI test, screenshot validation, XCUITest, snapshot test iOS | `ios-testing-tools` | Follow skill for Page Object, accessibility IDs, screenshots. |
| Android UI test, espresso, compose test, screenshot validation Android | `android-testing-tools` | Follow skill for test tags, Page Object, screenshots. |
| currency exchange, обмен валют, драмы, рубли, AMD, RUB, rate.am, раттам, курс обмена, поменять валюту, конвертация валют | `currency-exchange` | Use `exchange` CLI. NEVER scrape rate sources manually. |
| Go TUI test, bubbletea test, tuitestkit, go test helpers, reducer test, snapshot test go, golden file go, тесты для туи, го тест хелперы, тесты баблти, тест утилиты го, замкнутый цикл тестирования | `go-tui-testing-tools` | Follow skill for tuitestkit library, test patterns, templates, closed-loop workflow. |

---

## Workflow Details

For task-tracking or board-centric workflows, prefer the repo-local skill or the explicitly requested skill. Do not assume a global default board/task-management skill from infra alone.
