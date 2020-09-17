[Back to Python Menu](..)

# Train a model to predict likelihood of diabetes given some health parameters / observations

[Example Code: diabetes_xgboost.py](diabetes_xgboost.py)

## Steps of the Pipeline
 - Step1: Gather data and split it into training and validation sets
 - Step2: Fit the actual model
 - Step3: Run a set of predictions on the validation set. The function is designed to be more generic, it can be used to simply predict given a set of observations (dataset)
 - Step4: Calculate the accuracy score for the predictions

The workflow is designed for a dataset as defined in 
https://github.com/jbrownlee/Datasets/blob/master/pima-indians-diabetes.names

An example dataset is available at
https://raw.githubusercontent.com/jbrownlee/Datasets/master/pima-indians-diabetes.data.csv
