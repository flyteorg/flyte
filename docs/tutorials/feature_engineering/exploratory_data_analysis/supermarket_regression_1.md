---
celltoolbar: Tags
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
language_info:
  codemirror_mode:
    name: ipython
    version: 3
  file_extension: .py
  mimetype: text/x-python
  name: python
  nbconvert_exporter: python
  pygments_lexer: ipython3
  version: 3.9.9
vscode:
  interpreter:
    hash: 93d1c4f33f306e18e1c08a771c972fe86afbedaedb2338666e30a98a5179caac
---

# Supermarket Regression 1 Notebook

```{code-cell} ipython3
# reference: https://github.com/risenW/medium_tutorial_notebooks/blob/master/supermarket_regression.ipynb
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns

# makes graph display in notebook
%matplotlib inline
```

```{code-cell} ipython3
supermarket_data = pd.read_csv('https://raw.githubusercontent.com/risenW/medium_tutorial_notebooks/master/train.csv')
```

```{code-cell} ipython3
supermarket_data.head()
```

```{code-cell} ipython3
supermarket_data.describe()
```

```{code-cell} ipython3
# remove ID columns
cols_2_remove = ['Product_Identifier', 'Supermarket_Identifier', 'Product_Supermarket_Identifier']

newdata = supermarket_data.drop(cols_2_remove, axis=1)
```

```{code-cell} ipython3
newdata.head()
```

```{code-cell} ipython3
cat_cols = ['Product_Fat_Content','Product_Type',
            'Supermarket _Size', 'Supermarket_Location_Type',
           'Supermarket_Type' ]

num_cols = ['Product_Weight', 'Product_Shelf_Visibility',
            'Product_Price', 'Supermarket_Opening_Year', 'Product_Supermarket_Sales']
```

```{code-cell} ipython3
# bar plot for categorial features
for col in cat_cols:
    fig = plt.figure(figsize=(6,6)) # define plot area
    ax = fig.gca() # define axis  
    
    counts = newdata[col].value_counts() # find the counts for each unique category
    counts.plot.bar(ax = ax) # use the plot.bar method on the counts data frame
    ax.set_title('Bar plot for ' + col)
```

```{code-cell} ipython3
# scatter plot for numerical features
for col in num_cols:
    fig = plt.figure(figsize=(6,6)) # define plot area
    ax = fig.gca() # define axis  

    newdata.plot.scatter(x = col, y = 'Product_Supermarket_Sales', ax = ax)
```

```{code-cell} ipython3
# box plot for categorial features
for col in cat_cols:
    sns.boxplot(x=col, y='Product_Supermarket_Sales', data=newdata)
    plt.xlabel(col)
    plt.ylabel('Product Supermarket Sales')
    plt.show() 
```

```{code-cell} ipython3
# correlation matrix
corrmat = newdata.corr()
f,ax = plt.subplots(figsize=(5,4))
sns.heatmap(corrmat, square=True)
```

```{code-cell} ipython3
---
jupyter:
  outputs_hidden: true
---
# pair plot of columns without missing values
import warnings
warnings.filterwarnings('ignore')

cat_cols_pair = ['Product_Fat_Content','Product_Type','Supermarket_Location_Type']

cols_2_pair = ['Product_Fat_Content',
             'Product_Shelf_Visibility',
             'Product_Type',
             'Product_Price',
             'Supermarket_Opening_Year',
             'Supermarket_Location_Type',
             'Supermarket_Type',
             'Product_Supermarket_Sales']

for col in cat_cols_pair:
    sns.set()
    plt.figure()
    sns.pairplot(data=newdata[cols_2_pair], height=3.0, hue=col)
    plt.show()
```

```{code-cell} ipython3
# FEATURE ENGINEERING
# print all unique values
newdata['Product_Fat_Content'].unique()
```

```{code-cell} ipython3
fat_content_dict = {'Low Fat': 0, 'Ultra Low fat': 0, 'Normal Fat': 1}
newdata['is_normal_fat'] = newdata['Product_Fat_Content'].map(fat_content_dict)

# preview the values
newdata['is_normal_fat'].value_counts()
```

```{code-cell} ipython3
# assign year 2000 and above as 1, 1996 and below as 0
def cluster_open_year(year):
    if year <= 1996:
        return 0
    else:
        return 1
    
newdata['open_in_the_2000s'] = newdata['Supermarket_Opening_Year'].apply(cluster_open_year)
```

```{code-cell} ipython3
# preview feature
newdata[['Supermarket_Opening_Year', 'open_in_the_2000s']].head(4)
```

```{code-cell} ipython3
# get the unique categories in the column as a list
prod_type_cats = list(newdata['Product_Type'].unique())

# remove the class 1 categories
prod_type_cats.remove('Health and Hygiene')
prod_type_cats.remove('Household')
prod_type_cats.remove('Others')

def cluster_prod_type(product):
    if product in prod_type_cats:
        return 0
    else:
        return 1
    
newdata['Product_type_cluster'] = newdata['Product_Type'].apply(cluster_prod_type)
```

```{code-cell} ipython3
newdata[['Product_Type', 'Product_type_cluster']].tail(10)
```

```{code-cell} ipython3
# transforming skewed features
fig, ax = plt.subplots(1,2)

# plot of normal Product_Supermarket_Sales on the first axis
sns.histplot(data=newdata['Product_Supermarket_Sales'], bins=15, ax=ax[0])

# transform the Product_Supermarket_Sales and plot on the second axis
newdata['Product_Supermarket_Sales'] = np.log1p(newdata['Product_Supermarket_Sales'])
sns.histplot(data=newdata['Product_Supermarket_Sales'], bins=15, ax=ax[1])

plt.tight_layout()
plt.title("Transformation of Product_Supermarket_Sales feature")
```

```{code-cell} ipython3
# next, let's transform Product_Shelf_Visibility
fig, ax = plt.subplots(1,2)

# plot of normal Product_Supermarket_Sales on the first axis
sns.histplot(data=newdata['Product_Shelf_Visibility'], bins=15, ax=ax[0])

# transform the Product_Supermarket_Sales and plot on the second axis
newdata['Product_Shelf_Visibility'] = np.log1p(newdata['Product_Shelf_Visibility'])
sns.histplot(data=newdata['Product_Shelf_Visibility'], bins=15, ax=ax[1])

plt.tight_layout()
plt.title("Transformation of Product_Shelf_Visibility feature")
```

```{code-cell} ipython3
# feature encoding
for col in cat_cols:
    print('Value Count for', col)
    print(newdata[col].value_counts())
    print("---------------------------")
```

```{code-cell} ipython3
# save the target value to a new variable
import typing

# one hot encode using pandas dummy() function
dummified_data = pd.get_dummies(newdata)
dummified_data.head()
```

```{code-cell} ipython3
# fill-in missing values
# print null columns
dummified_data.isnull().sum()
```

```{code-cell} ipython3
# compute the mean
mean_pw = dummified_data['Product_Weight'].mean()

# fill the missing values with calculated mean
dummified_data['Product_Weight'].fillna(mean_pw, inplace=True)
```

```{code-cell} ipython3
# check if filling is successful
dummified_data.isnull().sum()
```

```{code-cell} ipython3
:tags: [outputs]

from flytekitplugins.papermill import record_outputs
record_outputs(dummified_data=dummified_data, dataset=dummified_data.to_json())
```

```{code-cell} ipython3

```
