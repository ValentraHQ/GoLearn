import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { AuthPage } from "./Auth";

const login = vi.fn();
const register = vi.fn();
vi.mock("@/lib/auth", () => ({ useAuth: () => ({ user: null, login, register }) }));
vi.mock("@/api/queries", () => ({ useConfig: () => ({ data: { auth: { local: true, oidc: false } } }) }));

function setup(mode: "login" | "register") {
  return render(
    <MemoryRouter initialEntries={["/x"]}>
      <Routes>
        <Route path="/x" element={<AuthPage mode={mode} />} />
        <Route path="/dashboard" element={<p>dashboard page</p>} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("AuthPage", () => {
  it("shows the server's error message and keeps the form usable", async () => {
    login.mockRejectedValueOnce(new Error("Invalid email or password."));
    setup("login");
    await userEvent.type(screen.getByLabelText("Email"), "ada@example.com");
    await userEvent.type(screen.getByLabelText("Password"), "wrong-password");
    await userEvent.click(screen.getByRole("button", { name: /sign in/i }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Invalid email or password.");
    expect(screen.getByRole("button", { name: /sign in/i })).toBeEnabled();
  });

  it("navigates to the dashboard after a successful registration", async () => {
    register.mockResolvedValueOnce(undefined);
    setup("register");
    await userEvent.type(screen.getByLabelText("Display name"), "Ada");
    await userEvent.type(screen.getByLabelText("Email"), "ada@example.com");
    await userEvent.type(screen.getByLabelText("Password"), "correct horse battery");
    await userEvent.click(screen.getByRole("button", { name: /create account/i }));
    await waitFor(() => expect(screen.getByText("dashboard page")).toBeInTheDocument());
    expect(register).toHaveBeenCalledWith("ada@example.com", "correct horse battery", "Ada");
  });

  it("is honest that OIDC isn't configured", () => {
    setup("login");
    expect(screen.getByText(/single sign-on \(oidc\) isn’t configured/i)).toBeInTheDocument();
  });
});
