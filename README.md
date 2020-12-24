# photon-dance-concurrent-dequeue

![dequeue](doc/dequeue.jpg)

### Usage

```golang
import (
    ...
    "github.com/amazingchow/photon-dance-concurrent-dequeue"
    ...
)

...
q := condequeue.NewConDequeue(500)
q.BPush("foo")
q.BPush("bar")

fmt.Println(q.Len())   // 2
fmt.Println(q.Front()) // "foo"
fmt.Println(q.Back())  // "bar"

q.FPop()  // remove "foo"
q.BPop()  // remove "bar"

q.FPush("hello")
q.BPush("world")

for q.Len() != 0 {
    fmt.Println(q.FPop())
}
...
```

## Contributing

### Step 1

* 🍴 Fork this repo!

### Step 2

* 🔨 HACK AWAY!

### Step 3

* 🔃 Create a new PR using https://github.com/amazingchow/photon-dance-concurrent-dequeue/compare!

## Support

* Reach out to me at <jianzhou42@163.com>.

## License

* This project is licensed under the MIT License - see the **[MIT license](http://opensource.org/licenses/mit-license.php)** for details.
