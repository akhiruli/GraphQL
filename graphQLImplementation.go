package main

import (
   "encoding/json"
   //"fmt"
   "log"
   "net/http"
   gql "github.com/graphql-go/graphql"
   "database/sql"
   _ "github.com/go-sql-driver/mysql"
   "io/ioutil"
   "context"
)

var (
   schema      gql.Schema
)

type Department struct{
   Deptno         int      `json:"deptno"`
   Dname          string   `json:"dname"`
   Loc            string   `json:"loc"`
}

type Employee struct{
   Empno          int      `json:"empno"`
   Ename          string   `json:"ename"`
   Job            string   `json:"job"`
   Mgr            int      `json:"mgr"`
   Hiredate       string   `json:"hiredate"`
   Sal            int      `json:"sal"`
   Com            int `json:"com"`
   Deptno         int      `json:"deptno"`
}

type Salgrade struct{
   Grade       int   `json:"grade"`
   Losal       int   `json:"losal"`
   Hisal       int   `json:"hisal"`
}

type ReturnT struct{
   Code        int      `json:"code"`
   Msg         string   `json:"msg"`
}

type DatabaseT struct{
   db    *sql.DB
}

var DepartmentType = gql.NewObject(
   gql.ObjectConfig{
      Name: "Department",
      Fields: gql.Fields{
         "deptno":&gql.Field{
            Type: gql.Int,
         },
         "dname": &gql.Field{
            Type: gql.String,
         },
         "loc": &gql.Field{
            Type: gql.String,
         },
      },
   },
)

var EmployeeType = gql.NewObject(
   gql.ObjectConfig{
      Name: "Employee",
      Fields: gql.Fields{
         "empno":&gql.Field{
            Type: gql.NewNonNull(gql.Int),
          },
         "ename": &gql.Field{
            Type: gql.String,
         },
         "job":&gql.Field{
            Type: gql.String,
         },
         "mgr":&gql.Field{
            Type: gql.Int,
         },
         "hiredate": &gql.Field{
            Type: gql.String,
         },
         "sal":&gql.Field{
            Type: gql.Int,
         },
         "com":&gql.Field{
            Type: gql.Int,
         },
         "deptno":&gql.Field{
            Type: gql.Int,
         },
      },
   },
)

var SalgradeType = gql.NewObject(
   gql.ObjectConfig{
      Name: "Salgrade",
      Fields: gql.Fields{
         "grade":&gql.Field{
            Type: gql.Int,
         },
         "losal":&gql.Field{
            Type: gql.Int,
         },
         "hisal":&gql.Field{
            Type: gql.Int,
         },
      },
   },
)

var ReturnType = gql.NewObject(
   gql.ObjectConfig{
      Name: "ReturnT",
      Fields: gql.Fields{
         "code": &gql.Field{
            Type: gql.Int,
         },
         "msg": &gql.Field{
            Type: gql.String,
         },
      },
   },
)



func ConnectToDB() *sql.DB{
   db, err := sql.Open("mysql", "mysql:password@/graphql")
   if err != nil {
      panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
   }

   err = db.Ping()
   if err != nil {
      panic(err.Error()) // proper error handling instead of panic in your app
   }

   log.Println("connected to db successfully.....")

   return db
}

func (database *DatabaseT) HandleReq(w http.ResponseWriter, r *http.Request){
   /*result := gql.Do(gql.Params{
            Schema: schema,
            RequestString: r.URL.Query().Get("query"),
         })*/

   body, err := ioutil.ReadAll(r.Body)
   if err != nil{
      http.Error(w, "query missing", 300)
      return
   }

   result := gql.Do(gql.Params{
            Schema: schema,
            RequestString: string(body),
            Context: context.WithValue(context.Background(), "dbconn", database.db),
         })

   json.NewEncoder(w).Encode(result)
}

func GetEmpbyId(p gql.ResolveParams)(interface{}, error){
   id, ok := p.Args["id"].(int)
   if ok{
      dbcon := p.Context.Value("dbconn").(*sql.DB)
      var emp Employee
      //err := db.QueryRow("SELECT empno, ename, job, mgr, hiredate, sal, comm, deptno from emp where empno=?", id).
      //            Scan(&emp.Empno, &emp.Ename, &emp.Job, &emp.Mgr, &emp.Hiredate, &emp.Sal, &emp.Com, &emp.Deptno)
      err := dbcon.QueryRow("SELECT empno, ename, job, mgr, hiredate, sal, deptno from emp where empno=?", id).
                  Scan(&emp.Empno, &emp.Ename, &emp.Job, &emp.Mgr, &emp.Hiredate, &emp.Sal, &emp.Deptno)
      if err != nil && err != sql.ErrNoRows {
         log.Printf("READ request failed with error %s for employee id %d\n", err.Error(), id)
         return nil, nil
      }

      if err == sql.ErrNoRows{
         return nil, nil
      }

      return emp, nil
   }

   return nil, nil
}

func GetDeptByEmpno(p gql.ResolveParams)(interface{}, error){
   empno, ok := p.Args["empno"].(int)
   if ok {
      dbcon := p.Context.Value("dbconn").(*sql.DB)
      var dept Department
      err := dbcon.QueryRow("select deptno, dname, loc from dept where deptno=(select deptno from emp where empno=?)", empno).Scan(&dept.Deptno, &dept.Dname, &dept.Loc)
      if err != nil && err != sql.ErrNoRows {
         log.Printf("READ request failed with error %s for employee id %d\n", err.Error(), empno)
         return nil, nil
      }

      if err == sql.ErrNoRows{
         return nil, nil
      }

      return dept, nil
   }

   return nil, nil
}

func GetEmployees(p gql.ResolveParams)(interface{}, error){
   var rows *sql.Rows
   var err error
   max, ok := p.Args["max"].(int)
   dbcon := p.Context.Value("dbconn").(*sql.DB)

   if ok {
      rows, err = dbcon.Query("SELECT empno, ename, job, mgr, hiredate, sal, deptno from emp limit ?", max)
   }else{
      rows, err = dbcon.Query("SELECT empno, ename, job, mgr, hiredate, sal, deptno from emp")
   }

   var emps []Employee

   if err != nil {
      log.Printf("Failed to get the employee data for the reason %s\n", err.Error())
      return emps, nil
   }

   defer rows.Close()
   for rows.Next() {
      var emp Employee
      err = rows.Scan(&emp.Empno, &emp.Ename, &emp.Job, &emp.Mgr, &emp.Hiredate, &emp.Sal, &emp.Deptno)
      if err != nil {
         log.Printf("Failed to parse with reason %s\n", err.Error())
         break
      }

      emps = append(emps, emp)
   }

   return emps, nil
}

func GetFields() gql.Fields{
   fields := gql.Fields{
      "hello": &gql.Field{
         Type: gql.String,
         Resolve: func(p gql.ResolveParams) (interface{}, error){
            return "world", nil
         },
      },

      "employee": &gql.Field{
         Type: EmployeeType,
         Description: "Get employee By ID",
         // We can define arguments that allow us to
         // pick specific tutorials. In this case
         // we want to be able to specify the ID of the
         // tutorial we want to retrieve
         Args: gql.FieldConfigArgument{
            "id": &gql.ArgumentConfig{
               Type: gql.Int,
            },
         },

         Resolve: GetEmpbyId,
      },

      "department": &gql.Field{
         Type: DepartmentType,
         Description:"Get department info by empno",
         Args: gql.FieldConfigArgument{
            "empno": &gql.ArgumentConfig{
               Type: gql.Int,
            },
         },

         Resolve: GetDeptByEmpno,
      },
      "employeelist": &gql.Field{
         Type: gql.NewList(EmployeeType),
         Description: "Get list of employee with limit",
         Args: gql.FieldConfigArgument{
            "max":&gql.ArgumentConfig{
               Type: gql.Int,
            },
         },
         Resolve: GetEmployees,
      },
   }

   return fields
}

func CreateEmployee(params gql.ResolveParams)(interface{}, error){
   var emp Employee
   emp.Empno = params.Args["empno"].(int)
   if val, ok := params.Args["ename"]; ok {
      emp.Ename = val.(string)
   }

   if val, ok := params.Args["job"]; ok{
      emp.Job = val.(string)
   }

   if val, ok := params.Args["mgr"]; ok{
      emp.Mgr = val.(int)
   }

   if val, ok := params.Args["hiredate"]; ok{
      emp.Hiredate = val.(string)
   }

   if val, ok := params.Args["sal"]; ok{
      emp.Sal = val.(int)
   }

   if val, ok := params.Args["com"]; ok{
      emp.Com = val.(int)
   }

   if val, ok := params.Args["deptno"]; ok{
      emp.Deptno = val.(int)
   }

   dbcon := params.Context.Value("dbconn").(*sql.DB)
   var ret ReturnT

   insert, err := dbcon.Query("INSERT INTO emp(empno, ename, job, mgr, hiredate, sal, comm, deptno) VALUES (?, ?, ?, ?, STR_TO_DATE(?, '%M %d %Y'), ?, ?, ?)", emp.Empno, emp.Ename, emp.Job, emp.Mgr, emp.Hiredate, emp.Sal, emp.Com, emp.Deptno)
   if err != nil{
      log.Println(err.Error())
      ret.Msg = err.Error()
   }else{
      insert.Close()
      ret.Msg = "success"
   }

   ret.Code = 123
   return ret, nil
}

func UpdateEmployee(params gql.ResolveParams)(interface{}, error){
   var emp Employee
   emp.Empno = params.Args["empno"].(int)

   return nil, nil
}

func GetMutaionFields() gql.Fields{
   fields := gql.Fields{
      "createemployee": &gql.Field{
         //Type: EmployeeType,
         Type: ReturnType,
         Description: "Create a new employee",
         Args: gql.FieldConfigArgument{
            "empno":&gql.ArgumentConfig{
               Type: gql.Int,
            },
            "ename": &gql.ArgumentConfig{
               Type: gql.String,
            },
            "job":&gql.ArgumentConfig{
               Type: gql.String,
            },
            "mgr":&gql.ArgumentConfig{
               Type: gql.Int,
            },
            "hiredate": &gql.ArgumentConfig{
               Type: gql.String,
            },
            "sal":&gql.ArgumentConfig{
               Type: gql.Int,
            },
            "com":&gql.ArgumentConfig{
               Type: gql.Int,
            },
            "deptno":&gql.ArgumentConfig{
               Type: gql.Int,
            },
         },
         Resolve: CreateEmployee,
      },
      "updateemployee": &gql.Field{
         //Type: EmployeeType,
         Type: ReturnType,
         Description: "Create a new employee",
         Args: gql.FieldConfigArgument{
            "empno":&gql.ArgumentConfig{
               Type: gql.Int,
            },
         },
         Resolve: UpdateEmployee,
      },
   }
   return fields
}

func main(){
   var err error

   database := DatabaseT{db: ConnectToDB()}
   defer database.db.Close()
  // ConnectToDB()

   rootQuery := gql.ObjectConfig{Name: "RootQuery", Fields: GetFields()}
   rootMutaion := gql.NewObject(gql.ObjectConfig{Name: "Mutation", Fields: GetMutaionFields()})
   schemaConfig := gql.SchemaConfig{Query: gql.NewObject(rootQuery), Mutation: rootMutaion}
   //schemaConfig := gql.SchemaConfig{Query: gql.NewObject(rootQuery)}
   schema, err = gql.NewSchema(schemaConfig)

   if err != nil {
      log.Fatalf("failed to create new schema, error: %v", err)
   }

   http.HandleFunc("/graphql", database.HandleReq)

   http.ListenAndServe(":12345", nil)

}
