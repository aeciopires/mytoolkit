<!-- TOC -->

- [JSON Tree Viewer — CLI](#json-tree-viewer--cli)
  - [Example](#example)

<!-- TOC -->

# JSON Tree Viewer — CLI

```
mytoolkit json-tree --in <file|-> [--out <file|->]
```

## Example

```
$ echo '{"a":1,"b":[true,null]}' | mytoolkit json-tree
{
  "type": "object",
  "children": [
    {
      "key": "a",
      "type": "number",
      "value": "1"
    },
    {
      "key": "b",
      "type": "array",
      "children": [
        {
          "type": "bool",
          "value": true
        },
        {
          "type": "null"
        }
      ]
    }
  ]
}
```

See `mytoolkit json-tree --help` for the full flag list.
