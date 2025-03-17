using BlazorApp.Components;
using Microsoft.AspNetCore.Authentication.Cookies;
using Microsoft.AspNetCore.Identity;
using Zitadel.Authentication;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container.
builder.Services.AddRazorComponents()
    .AddInteractiveServerComponents();

builder.Services.AddAuthorization().AddAuthentication(ZitadelDefaults.AuthenticationScheme)
.AddZitadel(
    o =>
    {
        o.Authority = "http://localhost:8080/";
        o.ClientId = "311607148553502723";
        o.SignInScheme = IdentityConstants.ExternalScheme;
        o.ResponseType = "code";
        o.SaveTokens = true;
        o.RequireHttpsMetadata = false;
        o.UsePkce = true;
    }
)
.AddExternalCookie().Configure(
    o =>
    {
        o.Cookie.HttpOnly = false;
        o.Cookie.IsEssential = true;
        o.Cookie.SameSite = SameSiteMode.None;
        o.Cookie.SecurePolicy = CookieSecurePolicy.Always;
    }
);

builder.Services.ConfigureCookieOidcRefresh(CookieAuthenticationDefaults.AuthenticationScheme, ZitadelDefaults.AuthenticationScheme);

builder.Services.AddCascadingAuthenticationState();

builder.Services.AddHttpContextAccessor();

builder.Services.AddHttpClient();

var app = builder.Build();

app.UseDeveloperExceptionPage();

app.UseRouting();
app.UseCookiePolicy();

app.UseAntiforgery();

app.UseStaticFiles();
app.UseAuthentication();
app.UseAuthorization();


app.MapRazorComponents<App>()
    .AddInteractiveServerRenderMode();

app.Run();
