#!/bin/bash
# fix Docker dns issue for spark driver
myip=`ifconfig | sed -En 's/127.0.0.1//;s/.*inet (addr:)?(([0-9]*\.){3}[0-9]*).*/\2/p'`
sp=' '
rs=$myip$sp$DNS_LABEL
cp /etc/hosts ~/hosts.new
sed -i -e '1 s/^/'"$rs"'\n/' ~/hosts.new
cp -f ~/hosts.new /etc/hosts

# configure azure storage account if exists
if [[ -n $AZURE_STORAGE_ACCOUNT_NAME || -n $AZURE_STORAGE_ACCOUNT_KEY ]]; then
  cp -p /spark/conf/spark-env.sh.template /spark/conf/spark-env.sh

  touch /spark/conf/spark-env.sh
  
  echo 'export SPARK_CONF_DIR="/spark/conf"' >> /spark/conf/spark-env.sh
  echo 'export HADOOP_CONF_DIR="/spark/conf"' >> /spark/conf/spark-env.sh

  touch /spark/conf/core-site.xml

  echo '<?xml version="1.0" encoding="UTF-8"?>' >> /spark/conf/core-site.xml
  echo '<?xml-stylesheet type="text/xsl" href="configuration.xsl"?>' >> /spark/conf/core-site.xml
  echo '<configuration>' >> /spark/conf/core-site.xml
  echo '<property>' >> /spark/conf/core-site.xml
  echo "<name>fs.azure.account.key.$AZURE_STORAGE_ACCOUNT_NAME.blob.core.windows.net</name>" >> /spark/conf/core-site.xml
  echo "<value>$AZURE_STORAGE_ACCOUNT_KEY</value>" >> /spark/conf/core-site.xml
  echo '</property>' >> /spark/conf/core-site.xml
  echo '</configuration>' >> /spark/conf/core-site.xml
fi

# start legion server
./main &
sleep 2

# continue with spark
chmod +x /master.sh
/master.sh