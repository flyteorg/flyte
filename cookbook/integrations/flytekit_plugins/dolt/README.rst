Dolt: Data versioning
===============================================

The `DoltTable` plugin is a wrapper that moves data between
`pandas.DataFrame`'s at execution time and database tables at rest.

Installation
------------

The dolt plugin and dolt command line tool are required to run these examples:
.. code:: bash

   pip install flytekitplugins.dolt
   sudo bash -c 'curl -L https://github.com/dolthub/dolt/releases/latest/download/install.sh | sudo bash'

Dolt requires a user configuration to run `init`:
.. code:: bash

   dolt config --global --add user.email <email>
   dolt config --global --add user.name <name>

These demoes assume a `foo` database has been created locally:
.. code:: bash

   mkdir foo
   cd foo
   dolt init


Quick Start
-----------

.. testcode:: dolt-quickstart

   doltdb_path = os.path.join(os.path.dirname(__file__), "foo")

   user_conf = DoltConfig(
       db_path=doltdb_path,
       tablename="users",
   )

   @task
   def populate_users(a: int) -> DoltTable:
       users = [("George", a), ("Alice", a*2), ("Stephanie", a*3)]
       df = pd.DataFrame(users, columns=["name", "count"])
       return DoltTable(data=df, config=user_conf)

   @task
   def count_users(table: DoltTable) -> pandas.DataFrame:
       return table.data

   @workflow
   def wf(a: int) -> pandas.DataFrame:
       users = populate_users(a=a)
       return users

   print(wf(a=2))

.. testoutput:: dolt-quickstart
   :options: +NORMALIZE_WHITESPACE

              name  count
   0  Stephanie     21
   1     George      7
   2      Alice     14
