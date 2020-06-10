<!--
title: Zeebe Workflow
weight: 4705
-->
# Zeebe Workflow

**This plugin is in ALPHA stage**

This activity allows you to create or cancel Zeebe Workflow instance.

## Installation

### Flogo CLI
```bash
flogo install github.com/codelity-co/flogo-zeebeworkflow-activity
```

## Configuration

### Settings:
  | Name                | Type   | Description
  | :---                | :---   | :---
  | zeebeBrokerHost     | string | Zeebe broker host - ***REQUIRED***
  | zeebeBrokerPort     | int    | Zeebe broker port, default 26500 - ***REQUIRED***
  | bpmnProcessID       | string | BPMN process ID - ***REQUIRED***
  | command             | string | Zeebe command, Create or Cancel - ***REQUIRED***

### Input
  | Name                | Type   | Description
  | :---                | :---   | :---
  | data                | object | data object - ***REQUIRED***

### Output:
  | Name          | Type   | Description
  | :---          | :---   | :---
  | status        | string | status text, ERROR or SUCCESS - ***REQUIRED***
  | result        | any    | activity result

## Example

```json
{
  "id": "flogo-cockroachdb-activity",
  "name": "Codelity Flogo CockroachDB Activity",
  "ref": "github.com/codelity-co/flogo-cockroachdb-activity",
  "settings": {
    "zeebeBrokerHost": "localhost",
    "zeebeBrokerPort": 26500,
    "bpmnProcessID": "order-process",
    "command": "Create"
  },
  "input": {
    "data": "=json.path(\"$.somepattern\", coerce.toObject($flow.dataobject))"
  }
}
```