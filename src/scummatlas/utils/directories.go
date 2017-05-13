 package utils

 import (
   "os"
   "fmt"
   "io"
 )

 func CopyFile(source string, dest string) (err error) {
     sourcefile, err := os.Open(source)
     if err != nil {
         return err
     }

     defer sourcefile.Close()
     destfile, err := os.Create(dest)
     if err != nil {
         return err
     }
     defer destfile.Close()

     _, err = io.Copy(destfile, sourcefile)
     if err == nil {
         sourceinfo, err := os.Stat(source)
         if err != nil {
             err = os.Chmod(dest, sourceinfo.Mode())
         }
     }

     return
 }


func CopyDir(source string, dest string) (err error) {
     sourceinfo, err := os.Stat(source)
     if err != nil {
         return err
     }

     err = os.MkdirAll(dest, sourceinfo.Mode())
     if err != nil {
         return err
     }

     directory, _ := os.Open(source)
     objects, err := directory.Readdir(-1)

     for _, obj := range objects {
         sourcefilepointer := source + "/" + obj.Name()
         destinationfilepointer := dest + "/" + obj.Name()

         if obj.IsDir() {
             err = CopyDir(sourcefilepointer, destinationfilepointer)
             if err != nil {
                 fmt.Println(err)
             }
         } else {
             err = CopyFile(sourcefilepointer, destinationfilepointer)
             if err != nil {
                 fmt.Println(err)
             }
         }
     }
     return
 }
