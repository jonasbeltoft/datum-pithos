using BlazorApp.Components;
using BlazorApp.Services;
using Blazored.LocalStorage;
using Blazored.LocalStorage.Serialization;
using Microsoft.AspNetCore.Authentication;
using Microsoft.AspNetCore.Components.Authorization;
using Microsoft.Extensions.DependencyInjection.Extensions;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container.
builder.Services.AddRazorComponents()
    .AddInteractiveServerComponents();

// api url for backend
var backendUrl = builder.Configuration.GetValue<string>("BACKEND_URL") ?? "http://localhost:8000/api/v1/";

builder.Services.AddScoped(sp => new HttpClient { BaseAddress = new Uri(backendUrl) });

builder.Services.AddCascadingAuthenticationState();
builder.Services.AddAuthorizationCore();

builder.Services.AddScoped<AuthenticationStateProvider, CustomAuthStateProvider>();
builder.Services.AddScoped<BrowserStorageService>();
builder.Services.AddScoped<MeasurementService>();

var app = builder.Build();

app.UseDeveloperExceptionPage();

app.UseRouting();
app.UseStaticFiles();
app.UseAntiforgery();

app.UseAuthorization();

app.MapRazorComponents<App>()
    .AddInteractiveServerRenderMode().AllowAnonymous();

// app.MapFallbackToPage("/");
app.UseStatusCodePagesWithRedirects("/404");

app.Run();
