# ast-index Rules

## Mandatory Search Rules

1. **ALWAYS use ast-index FIRST** for any code search task
2. **NEVER duplicate results** - if ast-index found usages/implementations,
   that IS the complete answer
3. **DO NOT run grep "for completeness"** after ast-index returns results
4. **Use grep/Search ONLY when:**
   - ast-index returns empty results
   - searching for regex patterns (ast-index uses literal match)
   - searching for string literals inside code (`"some text"`)
   - searching in comments content

## Why ast-index

ast-index is much faster than grep on large repos and returns structured,
accurate results.

## Common Command Reference

| Task | Command |
|------|---------|
| Universal search | `ast-index search "query"` |
| Find type/class | `ast-index class "Name"` |
| Find symbol | `ast-index symbol "Name"` |
| Find usages | `ast-index usages "Name"` |
| Find implementations | `ast-index implementations "Interface"` |
| Call hierarchy | `ast-index call-tree "function" --depth 3` |
| Find callers | `ast-index callers "functionName"` |
| Module deps | `ast-index deps "module-name"` |
| File outline | `ast-index outline "path/to/file.ext"` |
| File imports | `ast-index imports "path/to/file.ext"` |

## Index Management

- `ast-index rebuild` - Full reindex (run once after clone)
- `ast-index update` - After git pull/merge
- `ast-index stats` - Show index statistics

## Go Notes

- Exported functions, structs, and interfaces: `ast-index symbol "ShellSort"`
  or `ast-index class "MyStruct"`.
- Interface implementations: `ast-index implementations "sort.Interface"`.
- Who calls a function before changing its signature:
  `ast-index callers "selectStepSedgewick"`.
- Package-level layout of a file: `ast-index outline "lib/sort.go"`.
- Package imports of a file: `ast-index imports "main.go"`.
- Test helpers and benchmarks are regular Go symbols - search them the same way
  (`ast-index usages "IntSliceBig"`), do not fall back to grep for them.
