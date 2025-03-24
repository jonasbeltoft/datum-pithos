namespace BlazorApp.Services
{
	using System.Net;
	using System.Runtime.Serialization.Formatters.Binary;
	using System.Security.Claims;
	using System.Text.Json;
	using System.Text.Json.Nodes;
	using Microsoft.AspNetCore.Authentication;
	using Microsoft.AspNetCore.Authentication.Cookies;
	using Microsoft.AspNetCore.Components.Authorization;
	using Microsoft.AspNetCore.Http;
	using System.Net.Http.Headers;
	using Blazored.LocalStorage;
	using Microsoft.JSInterop;

	public class CustomAuthStateProvider : AuthenticationStateProvider
	{
		private readonly HttpClient httpClient;
		private readonly BrowserStorageService localStorage;

		public CustomAuthStateProvider(HttpClient _httpClient, BrowserStorageService _localStorage)
		{
			this.httpClient = _httpClient;
			this.localStorage = _localStorage;
		}

		public override async Task<AuthenticationState> GetAuthenticationStateAsync()
		{
			var user = new ClaimsPrincipal(new ClaimsIdentity());

			var accessToken = await localStorage.GetItemAsync("accessToken");
			if (accessToken != null)
			{
				httpClient.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", accessToken);
			}
			try
			{

				var response = await httpClient.GetAsync("profile");
				if (response.IsSuccessStatusCode)
				{
					var strResponse = await response.Content.ReadAsStringAsync();
					var jsonResponse = JsonNode.Parse(strResponse);
					var username = jsonResponse!["username"]!.ToString();
					var displayName = jsonResponse!["displayName"]!.ToString();
					var role = jsonResponse!["role"]!.ToString();

					var claims = new List<Claim> {
						new Claim("username", username),
						new Claim(ClaimTypes.Name, displayName),
						new Claim(ClaimTypes.Role, role),
					};

					var identity = new ClaimsIdentity(claims, "Token");
					user = new ClaimsPrincipal(identity);
					return new AuthenticationState(user);
				}
			}
			catch
			{
			}

			return new AuthenticationState(user);
		}

		public async void LogoutAsync()
		{
			// Logout server side
			await httpClient.PostAsync("logout", null);

			// Logout client side
			httpClient.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer");
			await localStorage.RemoveItemAsync("accessToken");
			await localStorage.RemoveItemAsync("refreshToken");

			NotifyAuthenticationStateChanged(GetAuthenticationStateAsync());
		}
		public async Task<LoginResult> LoginAsync(string username, string password)
		{
			try
			{
				var response = await httpClient.PostAsJsonAsync("login", new
				{
					username,
					password
				});

				if (response.IsSuccessStatusCode)
				{
					var strResponse = await response.Content.ReadAsStringAsync();
					var jsonResponse = JsonNode.Parse(strResponse);
					var accessToken = jsonResponse?["accessToken"]?.ToString();
					var refreshToken = jsonResponse?["refreshToken"]?.ToString();

					await localStorage.SetItemAsync("accessToken", accessToken);
					await localStorage.SetItemAsync("refreshToken", refreshToken);

					httpClient.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", accessToken);

					// Refresh auth state
					NotifyAuthenticationStateChanged(GetAuthenticationStateAsync());

					return new LoginResult { Succeeded = true };
				}
				else
				{
					return new LoginResult { Succeeded = false, Errors = ["Wrong email or password"] };
				}
			}
			catch { }

			return new LoginResult { Succeeded = false, Errors = ["Connection error"] };
		}
	}

	public class LoginResult
	{
		public bool Succeeded
		{
			get; set;
		}
		public string[] Errors
		{
			get; set;
		} = [];
	}

	public class BrowserStorageService
	{
		private readonly IJSRuntime jSRuntime;

		public BrowserStorageService(IJSRuntime jSRuntime)
		{
			this.jSRuntime = jSRuntime;
		}

		public async Task SetItemAsync(string key, string value)
		{
			await jSRuntime.InvokeVoidAsync("sessionStorage.setItem", key, value);
		}
		public async Task<string?> GetItemAsync(string key)
		{
			return await jSRuntime.InvokeAsync<string>("sessionStorage.getItem", key);
		}

		public async Task RemoveItemAsync(string key)
		{
			await jSRuntime.InvokeVoidAsync("sessionStorage.removeItem", key);
		}
	}
}
