package bootstrap

import (
	"html/template"
	"strings"
)

// GetTemplates load templates with their corresponding layouts and stores
// them in a map
func GetTemplates() map[string]*template.Template {
	templates := make(map[string]*template.Template)

	templates["home#index"] = loadTemplate("views/index.html")
	templates["home#unsupported"] = loadBlankTemplate("views/unsupported.html")

	templates["info#about"] = loadTemplate("views/info/about.html")
	templates["info#contact"] = loadTemplate("views/info/contact.html")
	templates["info#help"] = loadTemplate("views/info/help.html")
	templates["info#tos"] = loadTemplate("views/info/tos.html")

	templates["listing#create"] = loadTemplate("views/listing/create.html")
	templates["listing#edit"] = loadTemplate("views/listing/edit.html")
	templates["listing#section"] = loadTemplate("views/listing/section.html")
	templates["listing#selling"] = loadTemplate("views/listing/selling.html")
	templates["listing#view"] = loadTemplate("views/listing/view.html")

	templates["message#client"] = loadTemplate("views/message/client.html")

	templates["offer#buyer"] = loadTemplate("views/offer/buyer.html")
	templates["offer#buying"] = loadTemplate("views/offer/buying.html")
	templates["offer#seller"] = loadTemplate("views/offer/seller.html")

	templates["recover#index"] = loadTemplate("views/recover/index.html")
	templates["recover#reset"] = loadTemplate("views/recover/reset.html")

	templates["search#search"] = loadTemplate("views/search/search.html")

	templates["user#login"] = loadTemplate("views/user/login.html")
	templates["user#profile"] = loadTemplate("views/user/profile.html")
	templates["user#register"] = loadTemplate("views/user/register.html")

	templates["email#default"] = loadEmailTemplate("views/layouts/email.html")
	templates["email#plain"] = loadEmailTemplate("views/layouts/email_plain.html")

	return templates
}

func loadBlankTemplate(fpath string) *template.Template {

	funcs := template.FuncMap{
		"title":   strings.Title,
		"compare": strings.Compare,
	}

	return template.Must(
		template.New("").Funcs(funcs).ParseFiles(
			fpath))
}

func loadTemplate(fpath string) *template.Template {

	funcs := template.FuncMap{
		"title":   strings.Title,
		"compare": strings.Compare,
	}

	return template.Must(
		template.New("").Funcs(funcs).ParseFiles(
			"views/layouts/default.html",
			"views/layouts/sidebar.html",
			fpath))
}

func emailHelperIsEven(i int) bool {
	return i%2 == 0
}

func loadEmailTemplate(fpath string) *template.Template {
	funcs := template.FuncMap{
		"even": emailHelperIsEven,
	}

	return template.Must(
		template.New("").Funcs(funcs).ParseFiles(
			fpath))
}
