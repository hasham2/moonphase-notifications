package hello

import (
  "appengine"
  "appengine/datastore"
  "appengine/user"
  "appengine/mail"
  "fmt"
  "time"
  "html/template"
  "net/http"
)

type Contact struct {
  Email      string
  DateAdded  time.Time
}

func init(){
  http.HandleFunc("/", root)
  http.HandleFunc("/sign", sign)
  http.HandleFunc("/send", sendmail)
}

func root(w http.ResponseWriter, r *http.Request){
  fmt.Fprintf(w, guestbookForm)
}

const guestbookForm = `
<html>
  <body>
    <form action="/sign" method="post">
      <div><input type="text" name="content" width="40"/></div>
      <div><input type="submit" value="Keep me posted"></div>
    </form>
  </body>
</html>
`
func sign(w http.ResponseWriter, r *http.Request){
  c := appengine.NewContext(r)

  cl := Contact{
    Email: r.FormValue("content"),
    DateAdded: time.Now(),
  }

  key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "contact", nil), &cl)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  
  var e2 Contact
  if err = datastore.Get(c, key, &e2); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  err = signTemplate.Execute(w, e2.Email)

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

var signTemplate = template.Must(template.New("sign").Parse(signTemplateHTML))

const signTemplateHTML = `
<html>
  <body>
    <p>Thanks for providing your email we will keep you posted on latest updates:</p>
    <pre>{{.}}</pre>
  </body>
</html>
`

func handler(w http.ResponseWriter, r *http.Request){
  c := appengine.NewContext(r)
  u := user.Current(c)
  if u == nil {
    url, err := user.LoginURL(c, r.URL.String())
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    w.Header().Set("Location",url)
    w.WriteHeader(http.StatusFound)
    return 
  }
  fmt.Fprintf(w,"Hello, %v!", u)
}

func sendmail(w http.ResponseWriter, r *http.Request){

  c := appengine.NewContext(r)
  
  //calculate the moon phase start
  t := time.Now()
  var g,e int
  
  month := t.Month()
  year := t.Year()
  day := t.Day()

  if month == 1 {
    day--
  }else if month == 2 {
    day += 30
  }else{
    day += 28 + (int(month) - 2) * 3059/100
    // adjust for leap years
    ch := year & 3
    if ch != 0 {
      day++
    }
    if (year % 100) == 0 {
      day--
    }
  }
  g = (year - 1900) % 19 + 1
  e = (11 * g + 18) % 30
  if (e == 25 && g > 11) || e == 24 {
    e++
  }
  b := (((e + day) * 6 + 11) % 177) / 22 & 7
  //end moon phase calculation
  if b == 3 {
    q := datastore.NewQuery("Contact").Order("DateAdded")
    var contacts []Contact
    _, err := q.GetAll(c, &contacts)
    for _, p := range contacts {
        msg := &mail.Message{
        Sender: "Go Islamic Support <support@goislamic.com>",
        To: []string{p.Email},
        Subject: "You have new notifications waiting for you",
        Body: fmt.Sprintf(notificationMessage, p.Email),
        }
        if err = mail.Send(c, msg); err != nil {
          c.Errorf("Cannot send email: %v",err)
        }
     }
  }
}

const notificationMessage = `
Thanks for signing up for updates, You have been notified of coming full moon phase
Be aware !!!
%s
`
