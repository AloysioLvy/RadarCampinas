export default function Footer() {
    return (
      <footer className="bg-gray-100 p-4 border-t">
        <div className="container mx-auto text-center text-gray-600 text-sm">
          &copy; {new Date().getFullYear()} Radar Campinas - Todos os direitos reservados
        </div>
      </footer>
    );
  }
  