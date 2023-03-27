from pydriller import Repository
import sys

repo = "/work/repo"

changedlines = []
out = []
#out.append(('#Commit-Hash','file','type','expression','constants'))

for commit in Repository(repo, single=sys.argv[1]).traverse_commits():
    print("Parents are:")
    print(commit.parents)
    for file in commit.modified_files:
        if file.new_path is None:
            continue
        changes = []
        for added in file.diff_parsed["added"]:
            changes.append(added[0])
        changedlines.append((file.new_path, changes))

       

statsfile = open("/work/changed/cppstats_featurelocations.csv")

# FILENAME LINE_START LINE_END TYPE EXPRESSION CONSTANTS
# 0        1          2        3    4          5
featurelocations = [tuple(line.rstrip().split(",")) for line in statsfile]
featurelocations = featurelocations[2:] #remove headings and seperator lines at the top
for feature in featurelocations:
    #print(feature)
    for file in changedlines:
        #print(file)
        if feature[0].removesuffix(".xml").endswith(file[0]):
            #print("Processing: "+file[0])
            #correct file
            for line in file[1]:
                #print("is " + str(line) + " between " + feature[1] + " and " + feature[2] + "? ", end='')
                if line in range(int(feature[1]), int(feature[2])):
                    # Zeilen in Feature wurden geändert
                    out.append((sys.argv[1],file[0],feature[3],feature[4],",".join(feature[5].split(";"))))
                    break

statsfile.close()
print("===========================================")
save = open("/results/result.csv", "a")
for x in out:
    print(x)
    save.write(";".join(x)+"\n")
save.close()

import psycopg2

# Upload to Database
try:
    connection = psycopg2.connect(user="postgres",
                                  password="geheim",
                                  host="localhost",
                                  port="5432",
                                  database="Masterarbeit")
    cursor = connection.cursor()
    
    count = 0
    for x in out:
        # TODO: edit for right tablename!
        postgres_insert_query = """ INSERT INTO libxml2 (hash, file, type, expression, constants) VALUES (%s,%s,%s,%s,%s)"""
        record_to_insert = (x[0], x[1], x[2], x[3], x[4])  
        cursor.execute(postgres_insert_query, record_to_insert)
        count += cursor.rowcount

    connection.commit()
    print(count, "Record inserted successfully into postgres table")

except (Exception, psycopg2.Error) as error:
    print("Failed to insert record into postgres table", error)

finally:
    # closing database connection.
    if connection:
        cursor.close()
        connection.close()
        print("PostgreSQL connection is closed")