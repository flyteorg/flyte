# Map Tasks

If you have an array of inputs that you would like to run the same task over, Flyte provides a `dynamic_task` construct. Dynamic tasks are so named because they can produce additional tasks, the quantity and type of which is only known at run time, instead of regular Python tasks, which are all incorporated into workflows at compile time.

Continuing with the image rotation task, introduced in the [task](../task/) recipe, suppose we want to run the same task three times over threedifferent input images.

```python
images = [
    'https://upload.wikimedia.org/wikipedia/commons/a/a8/Fractal_pyramid.jpg',
    'https://upload.wikimedia.org/wikipedia/commons/thumb/2/21/Mandel_zoom_00_mandelbrot_set.jpg/640px-Mandel_zoom_00_mandelbrot_set.jpg',
    'https://upload.wikimedia.org/wikipedia/commons/thumb/a/ad/Julian_fractal.jpg/256px-Julian_fractal.jpg',
]
```

Instead of forcing users to run the task three times manually like,

```python
@workflow_class
class ManualCopy():
    img0 = rotate(images[0])
    img1 = rotate(images[1])
    img2 = rotate(images[2])
    
    out0 = Output(img0.outputs.out_image, sdk_type=Types.Blob)
    out1 = Output(img1.outputs.out_image, sdk_type=Types.Blob)
    out2 = Output(img2.outputs.out_image, sdk_type=Types.Blob)
```

we can use a `dynamic_task` to achieve a better result. Within a dynamic task, you can yield the regular `rotate` Python task.

```python
    results = []
    for ii in input_images:
        rotate_task = rotate(image_location=ii)
        yield rotate_task
        results.append(rotate_task.outputs.out_image)
```

Please see the attached [Python workflow](batch_rotate.py) for a full example. We've rewritten the rotate task here to do the image download in the task itself. Most installations of a Flyte sandbox are pretty resource limited, so we've tamped down the requests and limits pretty hard to ensure that all three subtasks get run.
