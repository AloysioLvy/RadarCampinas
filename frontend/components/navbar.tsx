
import Link from "next/link";

export default function Navbar() {
  return (
    <header className="bg-[#4d4dff] text-white p-4 shadow-md">
      <div className="container mx-auto flex justify-between items-center">
        <h1 className="text-2xl font-bold">Radar Campinas</h1>
        <nav>
          <ul className="flex space-x-4">
            <li>
              <Link href="/" className="hover:underline font-medium">
                Mapa
              </Link>
            </li>
            <li>
              <Link href="/chatbot" className="hover:underline font-medium">
                Reportar
              </Link>
            </li>
          </ul>
        </nav>
      </div>
    </header>
  );
}
