from email import message
from email.policy import default
from random import choices
import inquirer
import glob
import psycopg2

try:
    connection = psycopg2.connect(user="postgres", password="geheim", host="localhost",port="5432", database="Masterarbeit")
    cursor = connection.cursor()
    cursor.execute("select table_name from information_schema.tables where table_schema='public'")
    tables = cursor.fetchall()

except (Exception, psycopg2.Error) as error:
    print("Error while fetching data from PostgreSQL", error)

finally:
    # closing database connection.
    if connection:
        cursor.close()
        connection.close()
        print("PostgreSQL connection is closed")


# Select inputfile for analysis
questions = [
  inquirer.List('table',
                message="Which table should be processed?",
                choices=[t[0] for t in tables],
            ),
#  inquirer.List('hash',
#                message="Which column contains the hash?",
#                choices=[0,1,2,3,4,5],
#                default=1,
#            ),
#  inquirer.List('feature',
#                message="Which column contains the feature-list?",
#                choices=[0,1,2,3,4,5],
#            ),
#                default=5,
]
answers = inquirer.prompt(questions)
print("-------------------------------------------------------------------")

try:
    connection = psycopg2.connect(user="postgres", password="geheim", host="localhost",port="5432", database="Masterarbeit")
    cursor = connection.cursor()
    cursor.execute("select hash, constants from "+answers["table"])
    content = cursor.fetchall()
#    print(content)
    print("found " + str(len(content)) + "rows")
except (Exception, psycopg2.Error) as error:
    print("Error while fetching data from PostgreSQL", error)

finally:
    # closing database connection.
    if connection:
        cursor.close()
        connection.close()
        print("PostgreSQL connection is closed")





# Read file to array
res = {}
for row in content:
    hash = row[0]
    if hash in res:
        res[hash].extend(row[1].split(","))
    else:
        res[hash] = row[1].split(",")

# filter Features in Commit by using the datatype set 
distinct = {}
for key in res:
    distinct[key] = set(res[key])

# sort hashes bei number of changged features
s = dict(sorted(distinct.items(), key=lambda item: len(item[1])))

# print the 5 highest
print("The five commits with the highest number of features are:")
for k in {k: s[k] for k in list(s)[len(s)-5:]}:
    print(k+" - "+str(len(s[k])))
print("-------------------------------------------------------------------")

# count 
onlyone = 0
more = 0
for k in s:
    if len(s[k]) == 1:
        onlyone +=1
    else:
        more +=1

print("Anzahl Commits mit Änderungen in nur einem Feature:      "+str(onlyone))
print("Anzahl Commits mit Änderungen in mehr als einem Feature: "+str(more))
print("Anzahl Commits mit Änderungen an Featuren gesamt:        "+str(len(res)))
print("-------------------------------------------------------------------")

feat = {}
for k in s:
    for f in s[k]:
        if f in feat:
            feat[f].append(k)
        else:
            feat[f] = [k]
feat = dict(sorted(feat.items(), key=lambda item: len(item[1])))

print("The five features seen in the highest number of commits are:")
for k in {k: feat[k] for k in list(feat)[len(feat)-5:]}:
    print(f"{k:30} - "+str(len(feat[k])))
print("-------------------------------------------------------------------")

onlyone = 0
more = 0

for k in feat:
    if len(feat[k]) == 1:
        onlyone +=1
    else:
        more +=1
print("Anzahl Features in nur einem Commit:      "+str(onlyone))
print("Anzahl Features in mehr als einem Commit: "+str(more))
print("Anzahl aller Features:                    "+str(len(feat)))