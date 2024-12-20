ARG PYTHON_VERSION
FROM python:${PYTHON_VERSION}-slim-bookworm

WORKDIR /root
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8
ENV PYTHONPATH /root

ARG VERSION

RUN pip install sqlalchemy \
		psycopg2-binary \
		pymysql \
		flytekitplugins-sqlalchemy==$VERSION \
		flytekit==$VERSION

RUN useradd -u 1000 flytekit
RUN chown flytekit: /root
USER flytekit
