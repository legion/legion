from pyspark import SparkConf, SparkContext

from operator import add
import sys
## Constants
APP_NAME = "Meseeks Test"
##OTHER FUNCTIONS/CLASSES

def main(sc):
   somearr = ["Geez Rick", "I'm not sure..", "about this"]
   textRDD = sc.parallelize(somearr)
   words = textRDD.flatMap(lambda x: x.split(' ')).map(lambda x: (x, 1))
   wordcount = words.reduceByKey(add).collect()

   for wc in wordcount:
      print wc[0],wc[1]

   print "done"
 

if __name__ == "__main__":

   # Configure Spark
   conf = SparkConf().setAppName(APP_NAME)
   sc   = SparkContext(conf=conf)
   
   # Execute Main functionality
   main(sc)
