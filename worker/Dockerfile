FROM bde2020/spark-worker:2.3.0-hadoop2.7

RUN wget http://central.maven.org/maven2/org/apache/hadoop/hadoop-azure/2.7.0/hadoop-azure-2.7.0.jar
RUN wget http://central.maven.org/maven2/com/microsoft/azure/azure-storage/4.0.0/azure-storage-4.0.0.jar
RUN cp hadoop-azure-2.7.0.jar /spark/jars
RUN cp azure-storage-4.0.0.jar /spark/jars

RUN mkdir -p "/legionworker"

ADD . "/legionworker"
COPY setup_worker.sh "/legionworker"
WORKDIR "/legionworker"

CMD ["/bin/bash", "./setup_worker.sh"]
