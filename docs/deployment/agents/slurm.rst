.. _deployment-agent-setup-slurm:

Slurm agent
===========

This guide provides a comprehensive overview of setting up an environment to test the Slurm agent locally and enabling the agent in your Flyte deployment. Before proceeding, the first and foremost step is to spin up your own Slurm cluster, as it serves as the foundation for the setup. 

Spin up a Slurm cluster
-----------------------

Setting up a Slurm cluster can be challenging due to the limited detail in the `official instructions <https://slurm.schedmd.com/quickstart_admin.html#quick_start>`_. This tutorial simplifies the process, focusing on configuring a single-host Slurm cluster with ``slurmctld`` (central management daemon) and ``slurmd`` (compute node daemon).

Install MUNGE
~~~~~~~~~~~~~

.. epigraph::
  
  `MUNGE <https://dun.github.io/munge/>`_ is an authentication service, allowing a process to authenticate the UID and GID of another local or remote process within a group of hosts having common users and groups.

1. Install necessary packages
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: shell

  sudo apt install munge libmunge2 libmunge-dev

2. Generate and verify a MUNGE credential
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

After a MUNGE credential is generated, you can decode and verify the encoded token as follows:

.. code-block:: shell

  munge -n | unmunge | grep STATUS

.. note::

  A status of ``STATUS: Success(0)`` is expected and the MUNGE key is stored at ``/etc/munge/munge.key``. If the key is absent, please run the following command to create one manually:

  .. code:: shell

    sudo /usr/sbin/create-munge-key

3. Change the ownership and permissions of specific MUNGE-related directories
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

The ``munged`` daemon should run as the non-privileged "munge" user, which is created automatically. You can modify directory ownership and permissions as below:

.. code-block:: shell

  sudo chown -R munge: /etc/munge/ /var/log/munge/ /var/lib/munge/ /run/munge/
  sudo chmod 0700 /etc/munge/ /var/log/munge/ /var/lib/munge/
  sudo chmod 0755 /run/munge/
  sudo chmod 0700 /etc/munge/munge.key
  sudo chown -R munge: /etc/munge/munge.key

4. Start MUNGE
^^^^^^^^^^^^^^

Please make ``munged`` start at boot and restart the service:

.. code-block:: shell

  sudo systemctl enable munge
  sudo systemctl restart munge

.. note::

  To check if the daemon runs as expected, you can either use ``systemctl status munge`` or inspect the log file under ``/var/log/munge``.

Create a dedicated Slurm user 
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. epigraph::
  
  The *SlurmUser* must be created as needed prior to starting Slurm and must exist on all nodes in your cluster.

  -- `Slurm Super Quick Start <https://slurm.schedmd.com/quickstart_admin.html#quick_start>`_

Please make sure that ``uid`` is equal to ``gid`` to avoid some troubles: 

.. code-block:: shell

  adduser --system --uid <uid> --group --home /var/lib/slurm slurm

.. note::
 
  A system user usually has an ``uid`` in the range of 0-999, please refer to the section `Add a system user <https://manpages.ubuntu.com/manpages/oracular/en/man8/adduser.8.html>`_.

Once the system user is created, you can verify it using the following command:

.. code-block:: shell

  cat /etc/passwd | grep <uid>

It's of vital importance to set correct ownership of specific Slurm-related directories to prevent access issue. Directories mentioned below will be created automatically when the Slurm services start. However, manually creating them and altering the ownership beforehand help reduce errors:

Properly setting ownership of specific Slurm-related directories is crucial to avoid access issues. These directories are created automatically when Slurm services start, but manually creating them and adjusting ownership beforehand can make setup easier:

.. code-block:: shell

  sudo mkdir -p /var/spool/slurmctld /var/spool/slurmd /var/log/slurm
  sudo chown -R slurm: /var/spool/slurmctld /var/spool/slurmd /var/log/slurm

Run the Slurm cluster 
~~~~~~~~~~~~~~~~~~~~~

1. Install Slurm packages
^^^^^^^^^^^^^^^^^^^^^^^^^

First, you can download the Slurm source from `here <https://www.schedmd.com/download-slurm/>`_ (we'll use version ``24.05.5`` for illustration):

.. code-block:: shell

  wget -P <your-dir> https://download.schedmd.com/slurm/slurm-24.05.5.tar.bz2

.. note::

  We recommend to download the file to a clean directory because all Debian packages will be generate under this path. 

Then, Debian packages can be built following this `official guide <https://slurm.schedmd.com/quickstart_admin.html#debuild>`_:

.. code-block:: shell

  # Install basic Debian package build requirements
  sudo apt-get update
  sudo apt-get install build-essential fakeroot devscripts equivs

  # (Optional) Install dependencies if missing
  sudo apt install -y \
      libncurses-dev libgtk2.0-dev libpam0g-dev libperl-dev liblua5.3-dev \
      libhwloc-dev dh-exec librrd-dev libipmimonitoring-dev hdf5-helpers \
      libfreeipmi-dev libhdf5-dev man2html-base libcurl4-openssl-dev \
      libpmix-dev libhttp-parser-dev libyaml-dev libjson-c-dev \
      libjwt-dev liblz4-dev libmariadb-dev libdbus-1-dev librdkafka-dev

  # Unpack the distributed tarball
  tar -xaf slurm-24.05.5.tar.bz2

  # cd to the directory containing the Slurm source
  cd slurm-24.05.5

  # Install the Slurm package dependencies
  sudo mk-build-deps -i Debian/control

  # Build the Slurm packages
  debuild -b -uc -us

Debian packages are built and placed under the parent directory ``<your-dir>``. Since the single-host Slurm cluster functions as both a controller and a compute node, the following packages are required: ``slurm-smd``, ``slurm-smd-client`` (for CLI), ``slurm-smd-slurmctld``, and ``slurm-smd-slurmd``.

.. code-block:: shell

  # cd to the parent directory
  cd ..

  sudo dpkg -i slurm-smd_24.05.5-1_amd64.deb
  sudo dpkg -i slurm-smd-client_24.05.5-1_amd64.deb
  sudo dpkg -i slurm-smd-slurmctld_24.05.5-1_amd64.deb
  sudo dpkg -i slurm-smd-slurmd_24.05.5-1_amd64.deb

.. note::

  Please refer to `Installing Packages <https://slurm.schedmd.com/quickstart_admin.html#pkg_install>`_ for package selection.


2. Generate a Slurm configuration file 
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

After installation, generate a valid ``slurm.conf`` file for the Slurm cluster. We recommend using the `official configurator <https://slurm.schedmd.com/configurator.html>`_ to create it.

The following key-value pairs need to be set manually. Please leave the other options unchanged, as the default settings are sufficient for running ``slurmctld`` and ``slurmd``.

.. code-block:: ini 

  # == Cluster Name ==
  ClusterName=localcluster

  # == Control Machines ==
  SlurmctldHost=localhost

  # == Process Tracking ==
  ProctrackType=proctrack/linuxproc

  # == Event Logging ==
  SlurmctldLogFile=/var/log/slurm/slurmctld.log
  SlurmdLogFile=/var/log/slurm/slurmd.log

  # == Compute Nodes == 
  NodeName=localhost CPUs=16 RealMemory=30528 Sockets=1 CoresPerSocket=8 ThreadsPerCore=2 State=UNKNOWN
  PartitionName=debug Nodes=ALL Default=YES MaxTime=INFINITE State=UP

After completing the form, submit it, copy the content, and save it to ``/etc/slurm/slurm.conf``.

.. note::

  For a sample configuration file, please refer to this `slurm.conf <https://github.com/JiangJiaWei1103/Slurm-101/blob/main/slurm.conf>`_.

3. Start daemons
^^^^^^^^^^^^^^^^

Finally, enable ``slurmctld`` and ``slurmd`` to start at boot and restart them.

.. code-block:: shell

  # For controller
  sudo systemctl enable slurmctld
  sudo systemctl restart slurmctld

  # For compute
  sudo systemctl enable slurmd

You can verify the status of the daemons using ``systemctl status <daemon>`` or check the logs in ``/var/log/slurm/slurmctld.log`` and ``/var/log/slurm/slurmd.log`` to ensure the Slurm cluster is running correctly.





