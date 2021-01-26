from recipes.map_tasks.batch_rotate import BatchRotateWorkflow

raw_output_lp = BatchRotateWorkflow.create_launch_plan(
    raw_output_data_prefix='s3://my-s3-bucket/secondary-offloaded-location')
