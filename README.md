# OAM(Open AI Model)

`双兔傍地走，安能辨我是雄雌`

<img src="docs/oam.png" alt="示意图" width="30%" height="30%">

## Spec

```GO
type AI interface {
	SloveProblem(question string) (output interface{}, cost time.Duration)
}
```

The AI ​​that takes the shortest time is better.

Some may ask, why use time to measure the quality of answers?

Because **time will tell the truth.**

## Example

See [this](function/local/gorm/analyze/module.go).


## Differences between OAM（Open Application Model） and OAM（Open AI Model）

OAM（Open Application Model） is an open model for defining cloud native apps.

OAM（Open AI Model） is an open model for defining AI.

Focused on AI rather than application, Open AI Model [OAM] brings simplest but most powerful  design for modeling AI.


## Why does OAM fail?

They know nothing about DevOps.

The basic model is not right,the following content is completely wrong.

Actual DevOps job is like that:

![image](docs/suo.png)
