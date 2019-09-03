# GraphQL
A practical implementation of GraphQL in GO using the MySQL DB as the data source.

#need some packages
go get github.com/graphql-go/graphql
github.com/go-sql-driver/mysql

#sample query
curl -v -d '{employee(id:7369){empno,ename,deptno,hiredate}}' http://localhost:12345/graphql
curl -v -d 'mutation{createemployee(empno:7369,ename:"Jhonson",job:"SE",mgr:7782,sal:30000,com:3000,deptno:10,hiredate:"July 7 1987"){code}}' http://localhost:12345/graphql
